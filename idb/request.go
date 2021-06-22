// +build js,wasm

package idb

import (
	"log"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

var (
	jsIDBRequest = js.Global().Get("IDBRequest")
)

type Request struct {
	jsRequest js.Value
}

func wrapRequest(jsRequest js.Value) *Request {
	if !jsRequest.InstanceOf(jsIDBRequest) {
		panic("Invalid JS request type")
	}
	return &Request{jsRequest}
}

func (r *Request) Result() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return r.jsRequest.Get("result"), nil
}

func (r *Request) Err() (err error) {
	defer exception.Catch(&err)
	jsErr := r.jsRequest.Get("error")
	if jsErr.Truthy() {
		return js.Error{Value: jsErr}
	}
	return nil
}

func (r *Request) Await() (result js.Value, err error) {
	done := make(chan struct{})
	r.Listen(func() {
		result, err = r.Result()
		close(done)
	}, func() {
		err = r.Err()
		close(done)
	})
	<-done
	return
}

func (r *Request) transaction() *Transaction {
	return wrapTransaction(r.jsRequest.Get("transaction"))
}

func (r *Request) ListenSuccess(success func()) {
	r.Listen(success, nil)
}

func (r *Request) ListenError(failed func()) {
	r.Listen(nil, failed)
}

func (r *Request) Listen(success, failed func()) {
	panicHandler := func(err error) {
		log.Println("Failed resolving request results:", err)
		_ = r.transaction().Abort()
	}

	var errFunc, successFunc js.Func
	// setting up both is required to ensure boath are always released
	errFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer exception.CatchHandler(panicHandler)
		errFunc.Release()
		successFunc.Release()
		if failed != nil {
			failed()
		}
		return nil
	})
	successFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer exception.CatchHandler(panicHandler)
		errFunc.Release()
		successFunc.Release()
		if success != nil {
			success()
		}
		return nil
	})
	r.jsRequest.Call(addEventListener, "error", errFunc)
	r.jsRequest.Call(addEventListener, "success", successFunc)
}
