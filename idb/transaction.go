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

var (
	modeCache       jscache.Strings
	durabilityCache jscache.Strings
)

type TransactionMode int

const (
	TransactionReadOnly TransactionMode = iota
	TransactionReadWrite
)

func parseMode(s string) TransactionMode {
	switch s {
	case "readwrite":
		return TransactionReadWrite
	default:
		return TransactionReadOnly
	}
}

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

type TransactionDurability int

const (
	DurabilityDefault TransactionDurability = iota
	DurabilityRelaxed
	DurabilityStrict
)

func parseDurability(s string) TransactionDurability {
	switch s {
	case "relaxed":
		return DurabilityRelaxed
	case "strict":
		return DurabilityStrict
	default:
		return DurabilityDefault
	}
}

func (d TransactionDurability) String() string {
	switch d {
	case DurabilityRelaxed:
		return "relaxed"
	case DurabilityStrict:
		return "strict"
	default:
		return "default"
	}
}

func (d TransactionDurability) JSValue() js.Value {
	return durabilityCache.Value(d.String())
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

func (t *Transaction) Database() (_ *Database, err error) {
	defer exception.Catch(&err)
	return wrapDatabase(t.jsTransaction.Get("db")), nil
}

func (t *Transaction) Durability() (_ TransactionDurability, err error) {
	defer exception.Catch(&err)
	return parseDurability(t.jsTransaction.Get("durability").String()), nil
}

func (t *Transaction) Err() (err error) {
	defer exception.Catch(&err)
	jsErr := t.jsTransaction.Get("error")
	if jsErr.Truthy() {
		return js.Error{Value: jsErr}
	}
	return
}

func (t *Transaction) Mode() (_ TransactionMode, err error) {
	defer exception.Catch(&err)
	return parseMode(t.jsTransaction.Get("mode").String()), nil
}

func (t *Transaction) ObjectStoreNames() (_ []string, err error) {
	defer exception.Catch(&err)
	return stringsFromArray(t.jsTransaction.Get("objectStoreNames"))
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
	t.jsTransaction.Call(addEventListener, "error", errFunc)
	t.jsTransaction.Call(addEventListener, "complete", completeFunc)
	return prom
}
