//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"errors"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/jscache"
	"github.com/hack-pad/safejs"
)

var (
	supportsTransactionCommit = checkSupportsTransactionCommit()

	errNotInTransaction = errors.New("Not part of a transaction")
)

func checkSupportsTransactionCommit() bool {
	idbTransaction, err := safejs.Global().Get("IDBTransaction")
	if err != nil {
		return false
	}
	prototype, err := idbTransaction.Get("prototype")
	if err != nil {
		return false
	}
	commit, err := prototype.Get("commit")
	if err != nil {
		return false
	}
	supported, err := commit.Truthy()
	return supported && err == nil
}

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

func (m TransactionMode) jsValue() safejs.Value {
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

func (d TransactionDurability) jsValue() safejs.Value {
	return durabilityCache.Value(d.String())
}

// Transaction provides a static, asynchronous transaction on a database.
// All reading and writing of data is done within transactions. You use Database to start transactions,
// Transaction to set the mode of the transaction (e.g. is it TransactionReadOnly or TransactionReadWrite),
// and you access an ObjectStore to make a request. You can also use a Transaction object to abort transactions.
type Transaction struct {
	db            *Database
	jsTransaction safejs.Value
	objectStores  map[string]*ObjectStore
}

func wrapTransaction(db *Database, jsTransaction safejs.Value) *Transaction {
	return &Transaction{
		db:            db,
		jsTransaction: jsTransaction,
		objectStores:  make(map[string]*ObjectStore, 1),
	}
}

// Database returns the database connection with which this transaction is associated.
func (t *Transaction) Database() (*Database, error) {
	return t.db, nil
}

// Durability returns the durability hint the transaction was created with.
func (t *Transaction) Durability() (TransactionDurability, error) {
	durability, err := t.jsTransaction.Get("durability")
	if err != nil {
		return 0, err
	}
	durabilityString, err := durability.String()
	if err != nil {
		return 0, err
	}
	return parseDurability(durabilityString), nil
}

// Err returns an error indicating the type of error that occurred when there is an unsuccessful transaction. Returns nil if the transaction is not finished, is finished and successfully committed, or was aborted with Transaction.Abort().
func (t *Transaction) Err() error {
	jsErr, err := t.jsTransaction.Get("error")
	if err != nil {
		return err
	}
	if truthy, err := jsErr.Truthy(); err != nil {
		return err
	} else if truthy {
		return js.Error{Value: safejs.Unsafe(jsErr)}
	}
	return nil
}

// Abort rolls back all the changes to objects in the database associated with this transaction.
func (t *Transaction) Abort() error {
	_, err := t.jsTransaction.Call("abort")
	return err
}

// Mode returns the mode for isolating access to data in the object stores that are in the scope of the transaction. The default value is TransactionReadOnly.
func (t *Transaction) Mode() (TransactionMode, error) {
	mode, err := t.jsTransaction.Get("mode")
	if err != nil {
		return 0, err
	}
	modeStr, err := mode.String()
	return parseMode(modeStr), err
}

// ObjectStoreNames returns a list of the names of ObjectStores associated with the transaction.
func (t *Transaction) ObjectStoreNames() ([]string, error) {
	objectStoreNames, err := t.jsTransaction.Get("objectStoreNames")
	if err != nil {
		return nil, err
	}
	return stringsFromArray(objectStoreNames)
}

// ObjectStore returns an ObjectStore representing an object store that is part of the scope of this transaction.
func (t *Transaction) ObjectStore(name string) (*ObjectStore, error) {
	if store, ok := t.objectStores[name]; ok {
		return store, nil
	}
	jsObjectStore, err := t.jsTransaction.Call("objectStore", name)
	if err != nil {
		return nil, err
	}
	store := wrapObjectStore(t, jsObjectStore)
	t.objectStores[name] = store
	return store, nil
}

// Commit for an active transaction, commits the transaction. Note that this doesn't normally have to be called — a transaction will automatically commit when all outstanding requests have been satisfied and no new requests have been made. Commit() can be used to start the commit process without waiting for events from outstanding requests to be dispatched.
func (t *Transaction) Commit() error {
	if !supportsTransactionCommit {
		return nil
	}

	_, err := t.jsTransaction.Call("commit")
	return err
}

// Await waits for success or failure, then returns the results.
func (t *Transaction) Await(ctx context.Context) error {
	resolveCtx, cancel := context.WithCancel(ctx)

	var rejectValue error
	if err := t.addCancelingEventListener(resolveCtx, cancel, "abort", func() {
		rejectValue = errors.New("transaction aborted")
	}); err != nil {
		return err
	}

	if err := t.addCancelingEventListener(resolveCtx, cancel, "complete", func() {}); err != nil {
		return err
	}

	if err := t.addCancelingEventListener(resolveCtx, cancel, "error", func() {
		rejectValue = t.Err()
	}); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-resolveCtx.Done():
		return rejectValue
	}
}

// addCancelingEventListener adds an event listener for fn() and cleans it up when the context is canceled.
// The listener only runs if the context has not completed yet, then cancels it.
func (t *Transaction) addCancelingEventListener(ctx context.Context, cancel context.CancelFunc, eventName string, fn func()) error {
	jsFunc, err := safejs.FuncOf(func(safejs.Value, []safejs.Value) interface{} {
		select {
		case <-ctx.Done():
		default:
			fn()
			cancel()
		}
		return nil
	})
	if err != nil {
		return err
	}
	_, err = t.jsTransaction.Call(addEventListener, t.db.callStrings.Value(eventName), jsFunc)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_, _ = t.jsTransaction.Call(removeEventListener, t.db.callStrings.Value(eventName), jsFunc) // clean up on best-effort basis
		jsFunc.Release()
	}()
	return nil
}
