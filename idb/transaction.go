// +build js,wasm

package idb

import (
	"context"
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

// TransactionMode defines the mode for isolating access to data in the transaction's current object stores.
type TransactionMode int

const (
	// TransactionReadOnly allows data to be read but not changed.
	TransactionReadOnly TransactionMode = iota
	// TransactionReadWrite allows reading and writing of data in existing data stores to be changed.
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

// JSValue implements js.Wrapper
func (m TransactionMode) JSValue() js.Value {
	return modeCache.Value(m.String())
}

// TransactionDurability is a hint to the user agent of whether to prioritize performance or durability when committing a transaction.
type TransactionDurability int

const (
	// DurabilityDefault indicates the user agent should use its default durability behavior for the storage bucket. This is the default for transactions if not otherwise specified.
	DurabilityDefault TransactionDurability = iota
	// DurabilityRelaxed indicates the user agent may consider that the transaction has successfully committed as soon as all outstanding changes have been written to the operating system, without subsequent verification.
	DurabilityRelaxed
	// DurabilityStrict indicates the user agent may consider that the transaction has successfully committed only after verifying all outstanding changes have been successfully written to a persistent storage medium.
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

// JSValue implements js.Wrapper
func (d TransactionDurability) JSValue() js.Value {
	return durabilityCache.Value(d.String())
}

// Transaction provides a static, asynchronous transaction on a database.
// All reading and writing of data is done within transactions. You use Database to start transactions,
// Transaction to set the mode of the transaction (e.g. is it TransactionReadOnly or TransactionReadWrite),
// and you access an ObjectStore to make a request. You can also use a Transaction object to abort transactions.
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

// Database returns the database connection with which this transaction is associated.
func (t *Transaction) Database() (_ *Database, err error) {
	defer exception.Catch(&err)
	return wrapDatabase(t.jsTransaction.Get("db")), nil
}

// Durability returns the durability hint the transaction was created with.
func (t *Transaction) Durability() (_ TransactionDurability, err error) {
	defer exception.Catch(&err)
	return parseDurability(t.jsTransaction.Get("durability").String()), nil
}

// Err returns an error indicating the type of error that occurred when there is an unsuccessful transaction. Returns nil if the transaction is not finished, is finished and successfully committed, or was aborted with Transaction.Abort().
func (t *Transaction) Err() (err error) {
	defer exception.Catch(&err)
	jsErr := t.jsTransaction.Get("error")
	if jsErr.Truthy() {
		return js.Error{Value: jsErr}
	}
	return
}

// Abort rolls back all the changes to objects in the database associated with this transaction.
func (t *Transaction) Abort() (err error) {
	defer exception.Catch(&err)
	t.jsTransaction.Call("abort")
	return nil
}

// Mode returns the mode for isolating access to data in the object stores that are in the scope of the transaction. The default value is TransactionReadOnly.
func (t *Transaction) Mode() (_ TransactionMode, err error) {
	defer exception.Catch(&err)
	return parseMode(t.jsTransaction.Get("mode").String()), nil
}

// ObjectStoreNames returns a list of the names of ObjectStores associated with the transaction.
func (t *Transaction) ObjectStoreNames() (_ []string, err error) {
	defer exception.Catch(&err)
	return stringsFromArray(t.jsTransaction.Get("objectStoreNames"))
}

// ObjectStore returns an ObjectStore representing an object store that is part of the scope of this transaction.
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

// Commit for an active transaction, commits the transaction. Note that this doesn't normally have to be called — a transaction will automatically commit when all outstanding requests have been satisfied and no new requests have been made. Commit() can be used to start the commit process without waiting for events from outstanding requests to be dispatched.
func (t *Transaction) Commit() (err error) {
	if !supportsTransactionCommit {
		return nil
	}

	defer exception.Catch(&err)
	t.jsTransaction.Call("commit")
	return nil
}

// Await waits for success or failure, then returns the results.
func (t *Transaction) Await(ctx context.Context) error {
	_, err := t.prepareAwait(ctx).Await()
	return err
}

func (t *Transaction) prepareAwait(ctx context.Context) promise.Promise {
	resolve, reject, prom := promise.NewChan(ctx)

	errFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go reject(t.Err())
		return nil
	})
	completeFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go resolve(nil)
		return nil
	})
	t.jsTransaction.Call(addEventListener, "error", errFunc)
	t.jsTransaction.Call(addEventListener, "complete", completeFunc)

	go func() {
		<-ctx.Done()
		t.jsTransaction.Call(removeEventListener, "error", errFunc)
		t.jsTransaction.Call(removeEventListener, "complete", completeFunc)
		errFunc.Release()
		completeFunc.Release()
	}()
	return prom
}
