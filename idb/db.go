// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

// Database provides a connection to a database. You can use a Database object to open a transaction on your database then create, manipulate, and delete objects (data) in that database.
type Database struct {
	jsDB js.Value
}

func wrapDatabase(jsDB js.Value) *Database {
	return &Database{jsDB}
}

// Name returns the name of the connected database.
func (db *Database) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return db.jsDB.Get("name").String(), nil
}

// Version returns the version of the connected database.
func (db *Database) Version() (_ uint, err error) {
	defer exception.Catch(&err)
	return uint(db.jsDB.Get("version").Int()), nil
}

// ObjectStoreNames returns a list of the names of the object stores currently in the connected database.
func (db *Database) ObjectStoreNames() (_ []string, err error) {
	defer exception.Catch(&err)
	return stringsFromArray(db.jsDB.Get("objectStoreNames"))
}

// CreateObjectStore creates and returns a new object store or index.
func (db *Database) CreateObjectStore(name string, options ObjectStoreOptions) (_ *ObjectStore, err error) {
	defer exception.Catch(&err)
	jsOptions := map[string]interface{}{
		"autoIncrement": options.AutoIncrement,
	}
	if options.KeyPath != "" {
		jsOptions["keyPath"] = options.KeyPath
	}
	jsObjectStore := db.jsDB.Call("createObjectStore", name, jsOptions)
	return wrapObjectStore(jsObjectStore), nil
}

// DeleteObjectStore destroys the object store with the given name in the connected database, along with any indexes that reference it.
func (db *Database) DeleteObjectStore(name string) (err error) {
	defer exception.Catch(&err)
	db.jsDB.Call("deleteObjectStore", name)
	return nil
}

// Close closes the connection to a database.
func (db *Database) Close() (err error) {
	defer exception.Catch(&err)
	db.jsDB.Call("close")
	return nil
}

// Transaction returns a transaction object containing the Transaction.ObjectStore() method, which you can use to access your object store.
func (db *Database) Transaction(mode TransactionMode, objectStoreNames ...string) (_ *Transaction, err error) {
	return db.TransactionWithOptions(TransactionOptions{Mode: mode}, objectStoreNames...)
}

// TransactionOptions contains all available options for creating and starting a Transaction
type TransactionOptions struct {
	Mode       TransactionMode
	Durability TransactionDurability
}

// TransactionWithOptions returns a transaction object containing the Transaction.ObjectStore() method, which you can use to access your object store.
func (db *Database) TransactionWithOptions(options TransactionOptions, objectStoreNames ...string) (_ *Transaction, err error) {
	defer exception.Catch(&err)
	optionsMap := make(map[string]interface{})
	if options.Durability > 0 {
		optionsMap["durability"] = options.Durability
	}

	args := []interface{}{sliceFromStrings(objectStoreNames), options.Mode}
	if len(optionsMap) > 0 {
		args = append(args, optionsMap)
	}

	jsTxn := db.jsDB.Call("transaction", args...)
	return wrapTransaction(jsTxn), nil
}
