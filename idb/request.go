// +build js,wasm

package idb

import (
	"log"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

var (
	jsIDBRequest = js.Global().Get("IDBRequest")
	jsIDBIndex   = js.Global().Get("IDBIndex")
)

// Request provides access to results of asynchronous requests to databases and database objects
// using event listeners. Each reading and writing operation on a database is done using a request.
type Request struct {
	jsRequest js.Value
}

func wrapRequest(jsRequest js.Value) *Request {
	if !jsRequest.InstanceOf(jsIDBRequest) {
		panic("Invalid JS request type")
	}
	return &Request{jsRequest}
}

// Source returns the source of the request, such as an Index or an ObjectStore. If no source exists (such as when calling Factory.Open), it returns nil for both.
func (r *Request) Source() (objectStore *ObjectStore, index *Index, err error) {
	defer exception.Catch(&err)
	jsSource := r.jsRequest.Get("source")
	if jsSource.InstanceOf(jsObjectStore) {
		objectStore = wrapObjectStore(jsSource)
	} else if jsSource.InstanceOf(jsIDBIndex) {
		index = wrapIndex(jsSource)
	}
	return
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (r *Request) Result() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return r.jsRequest.Get("result"), nil
}

// Err returns an error in the event of an unsuccessful request, indicating what went wrong.
func (r *Request) Err() (err error) {
	defer exception.Catch(&err)
	jsErr := r.jsRequest.Get("error")
	if jsErr.Truthy() {
		return js.Error{Value: jsErr}
	}
	return nil
}

// Await waits for success or failure, then returns the results.
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

// ReadyState returns the state of the request. Every request starts in the pending state. The state changes to done when the request completes successfully or when an error occurs.
func (r *Request) ReadyState() (_ string, err error) {
	defer exception.Catch(&err)
	return r.jsRequest.Get("readyState").String(), nil
}

// Transaction returns the transaction for the request. This can return nil for certain requests, for example those returned from Factory.Open unless an upgrade is needed. (You're just connecting to a database, so there is no transaction to return).
func (r *Request) Transaction() (_ *Transaction, err error) {
	defer exception.Catch(&err)
	return r.transaction(), nil
}

func (r *Request) transaction() *Transaction {
	return wrapTransaction(r.jsRequest.Get("transaction"))
}

// ListenSuccess invokes the callback when the request succeeds
func (r *Request) ListenSuccess(success func()) {
	r.Listen(success, nil)
}

// ListenError invokes the callback when the request fails
func (r *Request) ListenError(failed func()) {
	r.Listen(nil, failed)
}

// Listen invokes the success callback when the request succeeds and failed when it fails.
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

// UintRequest is a Request that retrieves a uint result
type UintRequest struct {
	*Request
}

func newUintRequest(req *Request) *UintRequest {
	return &UintRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (u *UintRequest) Result() (uint, error) {
	result, err := u.Request.Result()
	if err != nil {
		return 0, err
	}
	return uint(result.Int()), nil
}

// Await waits for success or failure, then returns the results.
func (u *UintRequest) Await() (uint, error) {
	result, err := u.Request.Await()
	if err != nil {
		return 0, err
	}
	return uint(result.Int()), nil
}

// ArrayRequest is a Request that retrieves an array of js.Values
type ArrayRequest struct {
	*Request
}

func newArrayRequest(req *Request) *ArrayRequest {
	return &ArrayRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (a *ArrayRequest) Result() ([]js.Value, error) {
	result, err := a.Request.Result()
	if err != nil {
		return nil, err
	}
	var values []js.Value
	err = iterArray(result, func(i int, value js.Value) bool {
		values = append(values, value)
		return true
	})
	return values, err
}

// Await waits for success or failure, then returns the results.
func (a *ArrayRequest) Await() ([]js.Value, error) {
	result, err := a.Request.Await()
	if err != nil {
		return nil, err
	}
	var values []js.Value
	err = iterArray(result, func(i int, value js.Value) bool {
		values = append(values, value)
		return true
	})
	return values, err
}

// AckRequest is a Request that doesn't retrieve a value, only used to detect errors.
type AckRequest struct {
	*Request
}

func newAckRequest(req *Request) *AckRequest {
	return &AckRequest{req}
}

// Result is a no-op. This kind of request does not retrieve any data in the result.
func (a *AckRequest) Result() {} // no-op

// Await waits for success or failure, then returns the results.
func (a *AckRequest) Await() error {
	_, err := a.Request.Await()
	return err
}

// CursorRequest is a Request that retrieves a Cursor
type CursorRequest struct {
	*Request
}

func newCursorRequest(req *Request) *CursorRequest {
	return &CursorRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (c *CursorRequest) Result() (_ *Cursor, err error) {
	defer exception.Catch(&err)
	result, err := c.Request.Result()
	if err != nil {
		return nil, err
	}
	return wrapCursor(result), nil
}

// Await waits for success or failure, then returns the results.
func (c *CursorRequest) Await() (_ *Cursor, err error) {
	defer exception.Catch(&err)
	result, err := c.Request.Await()
	if err != nil {
		return nil, err
	}
	return wrapCursor(result), nil
}

// CursorWithValueRequest is a Request that retrieves a CursorWithValue
type CursorWithValueRequest struct {
	*Request
}

func newCursorWithValueRequest(req *Request) *CursorWithValueRequest {
	return &CursorWithValueRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (c *CursorWithValueRequest) Result() (_ *CursorWithValue, err error) {
	defer exception.Catch(&err)
	result, err := c.Request.Result()
	if err != nil {
		return nil, err
	}
	return wrapCursorWithValue(result), nil
}

// Await waits for success or failure, then returns the results.
func (c *CursorWithValueRequest) Await() (_ *CursorWithValue, err error) {
	defer exception.Catch(&err)
	result, err := c.Request.Await()
	if err != nil {
		return nil, err
	}
	return wrapCursorWithValue(result), nil
}
