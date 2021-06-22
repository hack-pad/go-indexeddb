// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

type Database struct {
	jsDB js.Value
}

func wrapDatabase(jsDB js.Value) *Database {
	return &Database{jsDB}
}

func (db *Database) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return db.jsDB.Get("name").String(), nil
}

func (db *Database) Version() (_ uint, err error) {
	defer exception.Catch(&err)
	return uint(db.jsDB.Get("version").Int()), nil
}

func (db *Database) ObjectStoreNames() (_ []string, err error) {
	defer exception.Catch(&err)
	return stringsFromArray(db.jsDB.Get("objectStoreNames"))
}

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

func (db *Database) DeleteObjectStore(name string) (err error) {
	defer exception.Catch(&err)
	db.jsDB.Call("deleteObjectStore", name)
	return nil
}

func (db *Database) Close() (err error) {
	defer exception.Catch(&err)
	db.jsDB.Call("close")
	return nil
}

func (db *Database) Transaction(mode TransactionMode, objectStoreNames ...string) (_ *Transaction, err error) {
	return db.TransactionWithOptions(TransactionOptions{Mode: mode}, objectStoreNames...)
}

type TransactionOptions struct {
	Mode       TransactionMode
	Durability TransactionDurability
}

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
