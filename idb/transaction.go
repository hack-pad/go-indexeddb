// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
	"github.com/hack-pad/go-indexeddb/idb/internal/jscache"
	"github.com/hack-pad/go-indexeddb/idb/internal/promise"
)

var (
	supportsTransactionCommit = js.Global().Get("IDBTransaction").Get("prototype").Get("commit").Truthy()
)

type TransactionMode int

const (
	TransactionReadOnly TransactionMode = iota
	TransactionReadWrite
)

var modeCache jscache.Strings

func (m TransactionMode) String() string {
	switch m {
	case TransactionReadWrite:
		return "readwrite"
	default:
		return "readonly"
	}
}

func (m TransactionMode) JSValue() js.Value {
	return modeCache.Value(m.String())
}

type Transaction struct {
	jsTransaction js.Value
	objectStores  map[string]*ObjectStore
}

func wrapTransaction(jsTransaction js.Value) *Transaction {
	return &Transaction{
		jsTransaction: jsTransaction,
		objectStores:  make(map[string]*ObjectStore),
	}
}

func (t *Transaction) Abort() (err error) {
	defer exception.Catch(&err)
	t.jsTransaction.Call("abort")
	return nil
}

func (t *Transaction) ObjectStore(name string) (_ *ObjectStore, err error) {
	if store, ok := t.objectStores[name]; ok {
		return store, nil
	}
	defer exception.Catch(&err)
	jsObjectStore := t.jsTransaction.Call("objectStore", name)
	store := wrapObjectStore(jsObjectStore)
	t.objectStores[name] = store
	return store, nil
}

func (t *Transaction) Commit() (err error) {
	if !supportsTransactionCommit {
		return nil
	}

	defer exception.Catch(&err)
	t.jsTransaction.Call("commit")
	return nil
}

func (t *Transaction) Await() error {
	_, err := t.prepareAwait().Await()
	return err
}

func (t *Transaction) prepareAwait() promise.Promise {
	resolve, reject, prom := promise.NewChan()

	var errFunc, completeFunc js.Func
	errFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		err := t.jsTransaction.Get("error")
		t.jsTransaction.Call("abort")
		go func() {
			errFunc.Release()
			completeFunc.Release()
			reject(err)
		}()
		return nil
	})
	completeFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			errFunc.Release()
			completeFunc.Release()
			resolve(nil)
		}()
		return nil
	})
	t.jsTransaction.Call("addEventListener", "error", errFunc)
	t.jsTransaction.Call("addEventListener", "complete", completeFunc)
	return prom
}
