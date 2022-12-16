// +build js,wasm

// nolint:errcheck
package idb_test

import (
	"context"
	"errors"
	"github.com/hack-pad/go-indexeddb/idb"
	"syscall/js"
	"time"
)

// dbTimeout is the global timeout for operations with the storage
// [context.Context].
const dbTimeout = time.Second

// NewContext builds a context for indexedDb operations.
func NewContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), dbTimeout)
}

var (
	Books = []string{
		"Hitchhiker's Guide to the Galaxy",
		"Leaves of Grass",
		"The Great Gatsby",
		"The Hobbit",
	}
)

func Example() {
	// Create the 'library' database, then create a 'books' object store during setup.
	// The setup func can also upgrade the database from older versions.
	ctx := context.Background()
	openRequest, _ := idb.Global().Open(ctx, "library", 1, func(db *idb.Database, oldVersion, newVersion uint) error {
		db.CreateObjectStore("books", idb.ObjectStoreOptions{})
		return nil
	})
	db, _ := openRequest.Await(ctx)

	{ // Store some books in the library database.
		txn, _ := db.Transaction(idb.TransactionReadWrite, "books")
		store, _ := txn.ObjectStore("books")
		for _, bookTitle := range Books {
			store.Add(js.ValueOf(bookTitle))
		}
		txn.Await(ctx)
	}

	{ // Iterate through the books and print their titles.
		txn, _ := db.Transaction(idb.TransactionReadOnly, "books")
		store, _ := txn.ObjectStore("books")
		cursorRequest, _ := store.OpenCursor(idb.CursorNext)
		cursorRequest.Iter(ctx, func(cursor *idb.CursorWithValue) error {
			value, _ := cursor.Value()
			println(value.String())
			return nil
		})
	}
}

// Get is a generic helper for getting values from the given [idb.ObjectStore].
// Only usable by primary key.
func Get(db *idb.Database, objectStoreName string, key js.Value) (js.Value, error) {

	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadOnly, objectStoreName)
	if err != nil {
		return js.Undefined(), err
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return js.Undefined(), err
	}

	// Perform the operation
	getRequest, err := store.Get(key)
	if err != nil {
		return js.Undefined(), err
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	resultObj, err := getRequest.Await(ctx)
	cancel()
	if err != nil {
		return js.Undefined(), err
	} else if resultObj.IsUndefined() {
		return js.Undefined(), errors.New("unable to get from ObjectStore: result is undefined")
	}
	return resultObj, nil
}

// GetIndex is a generic helper for getting values from the given
// [idb.ObjectStore] using the given [idb.Index].
func GetIndex(db *idb.Database, objectStoreName,
	indexName string, key js.Value) (js.Value, error) {

	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadOnly, objectStoreName)
	if err != nil {
		return js.Undefined(), err
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return js.Undefined(), err
	}
	idx, err := store.Index(indexName)
	if err != nil {
		return js.Undefined(), err
	}

	// Perform the operation
	getRequest, err := idx.Get(key)
	if err != nil {
		return js.Undefined(), err
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	resultObj, err := getRequest.Await(ctx)
	cancel()
	if err != nil {
		return js.Undefined(), err
	} else if resultObj.IsUndefined() {
		return js.Undefined(), errors.New("unable to get from ObjectStore: result is undefined")
	}
	return resultObj, nil
}

// Put is a generic helper for putting values into the given [idb.ObjectStore].
// Equivalent to insert if not exists else update.
func Put(db *idb.Database, objectStoreName string, value js.Value) (*idb.Request, error) {
	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadWrite, objectStoreName)
	if err != nil {
		return nil, err
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return nil, err
	}

	// Perform the operation
	request, err := store.Put(value)
	if err != nil {
		return nil, err
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	err = txn.Await(ctx)
	cancel()
	if err != nil {
		return nil, err
	}
	return request, nil
}

// Delete is a generic helper for removing values from the given [idb.ObjectStore].
// Only usable by primary key.
func Delete(db *idb.Database, objectStoreName string, key js.Value) error {
	// Prepare the Transaction
	txn, err := db.Transaction(idb.TransactionReadWrite, objectStoreName)
	if err != nil {
		return err
	}
	store, err := txn.ObjectStore(objectStoreName)
	if err != nil {
		return err
	}

	// Perform the operation
	_, err = store.Delete(key)
	if err != nil {
		return err
	}

	// Wait for the operation to return
	ctx, cancel := NewContext()
	err = txn.Await(ctx)
	cancel()
	if err != nil {
		return err
	}
	return nil
}

// DeleteIndex is a generic helper for removing values from the
// given [idb.ObjectStore] using the given [idb.Index]. Requires passing
// in the name of the primary key for the store.
func DeleteIndex(db *idb.Database, objectStoreName,
	indexName, pkeyName string, key js.Value) error {

	value, err := GetIndex(db, objectStoreName, indexName, key)
	if err != nil {
		return err
	}

	err = Delete(db, objectStoreName, value.Get(pkeyName))
	if err != nil {
		return err
	}
	return nil
}
