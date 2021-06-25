// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

// ObjectStoreOptions contains all available options for creating an ObjectStore
type ObjectStoreOptions struct {
	KeyPath       js.Value
	AutoIncrement bool
}

// ObjectStore represents an object store in a database. Records within an object store are sorted according to their keys. This sorting enables fast insertion, look-up, and ordered retrieval.
type ObjectStore struct {
	jsObjectStore js.Value
}

func wrapObjectStore(jsObjectStore js.Value) *ObjectStore {
	return &ObjectStore{jsObjectStore: jsObjectStore}
}

// IndexNames returns a list of the names of indexes on objects in this object store.
func (o *ObjectStore) IndexNames() (_ []string, err error) {
	defer exception.Catch(&err)
	return stringsFromArray(o.jsObjectStore.Get("indexNames"))
}

// KeyPath returns the key path of this object store. If this returns js.Null(), the application must provide a key for each modification operation.
func (o *ObjectStore) KeyPath() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return o.jsObjectStore.Get("keyPath"), nil
}

// Name returns the name of this object store.
func (o *ObjectStore) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return o.jsObjectStore.Get("name").String(), nil
}

// Transaction returns the Transaction object to which this object store belongs.
func (o *ObjectStore) Transaction() (_ *Transaction, err error) {
	defer exception.Catch(&err)
	return wrapTransaction(o.jsObjectStore.Get("transaction")), nil
}

// AutoIncrement returns the value of the auto increment flag for this object store.
func (o *ObjectStore) AutoIncrement() (_ bool, err error) {
	defer exception.Catch(&err)
	return o.jsObjectStore.Get("autoIncrement").Bool(), nil
}

// Add returns an AckRequest, and, in a separate thread, creates a structured clone of the value, and stores the cloned value in the object store. This is for adding new records to an object store.
func (o *ObjectStore) Add(value js.Value) (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("add", value))
	return newAckRequest(req), nil
}

// AddKey is the same as Add, but includes the key to use to identify the record.
func (o *ObjectStore) AddKey(key, value js.Value) (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("add", value, key))
	return newAckRequest(req), nil
}

// Clear returns an AckRequest, then clears this object store in a separate thread. This is for deleting all current records out of an object store.
func (o *ObjectStore) Clear() (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("clear"))
	return newAckRequest(req), nil
}

// Count returns a UintRequest, and, in a separate thread, returns the total number of records in the store.
func (o *ObjectStore) Count() (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("count"))
	return newUintRequest(req), nil
}

// CountKey returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided key.
func (o *ObjectStore) CountKey(key js.Value) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("count", key))
	return newUintRequest(req), nil
}

// CountRange returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided KeyRange.
func (o *ObjectStore) CountRange(keyRange *KeyRange) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("count", keyRange))
	return newUintRequest(req), nil
}

// CreateIndex creates a new index during a version upgrade, returning a new Index object in the connected database.
func (o *ObjectStore) CreateIndex(name string, keyPath js.Value, options IndexOptions) (index *Index, err error) {
	defer exception.Catch(&err)
	jsIndex := o.jsObjectStore.Call("createIndex", name, keyPath, map[string]interface{}{
		"unique":     options.Unique,
		"multiEntry": options.MultiEntry,
	})
	return wrapIndex(jsIndex), nil
}

// Delete returns an AckRequest, and, in a separate thread, deletes the store object selected by the specified key. This is for deleting individual records out of an object store.
func (o *ObjectStore) Delete(key js.Value) (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("delete", key))
	return newAckRequest(req), nil
}

// DeleteIndex destroys the specified index in the connected database, used during a version upgrade.
func (o *ObjectStore) DeleteIndex(name string) (err error) {
	defer exception.Catch(&err)
	o.jsObjectStore.Call("deleteIndex", name)
	return nil
}

// GetAllKeys returns an ArrayRequest that retrieves record keys for all objects in the object store.
func (o *ObjectStore) GetAllKeys() (_ *ArrayRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("getAllKeys"))
	return newArrayRequest(req), nil
}

// GetAllKeysRange returns an ArrayRequest that retrieves record keys for all objects in the object store matching the specified query. If maxCount is 0, retrieves all objects matching the query.
func (o *ObjectStore) GetAllKeysRange(query *KeyRange, maxCount uint) (_ *ArrayRequest, err error) {
	defer exception.Catch(&err)
	args := []interface{}{query}
	if maxCount > 0 {
		args = append(args, maxCount)
	}
	req := wrapRequest(o.jsObjectStore.Call("getAllKeys", args...))
	return newArrayRequest(req), nil
}

// Get returns a Request, and, in a separate thread, returns the store object store selected by the specified key. This is for retrieving specific records from an object store.
func (o *ObjectStore) Get(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("get", key)), nil
}

// GetKey returns a Request, and, in a separate thread retrieves and returns the record key for the object matching the specified parameter.
func (o *ObjectStore) GetKey(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("getKey", value)), nil
}

// Index opens an index from this object store after which it can, for example, be used to return a sequence of records sorted by that index using a cursor.
func (o *ObjectStore) Index(name string) (index *Index, err error) {
	defer exception.Catch(&err)
	jsIndex := o.jsObjectStore.Call("index", name)
	return wrapIndex(jsIndex), nil
}

// Put returns a Request, and, in a separate thread, creates a structured clone of the value, and stores the cloned value in the object store. This is for updating existing records in an object store when the transaction's mode is readwrite.
func (o *ObjectStore) Put(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("put", value)), nil
}

// PutKey is the same as Put, but includes the key to use to identify the record.
func (o *ObjectStore) PutKey(key, value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("put", value, key)), nil
}

// OpenCursor returns a CursorWithValueRequest, and, in a separate thread, returns a new CursorWithValue. Used for iterating through an object store by primary key with a cursor.
func (o *ObjectStore) OpenCursor(direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openCursor", js.Null(), direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenCursorKey is the same as OpenCursor, but opens a cursor over the given key instead.
func (o *ObjectStore) OpenCursorKey(key js.Value, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openCursor", key, direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenCursorRange is the same as OpenCursor, but opens a cursor over the given range instead.
func (o *ObjectStore) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openCursor", keyRange, direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenKeyCursor returns a CursorRequest, and, in a separate thread, returns a new Cursor. Used for iterating through all keys in an object store.
func (o *ObjectStore) OpenKeyCursor(direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openKeyCursor", js.Null(), direction.String()))
	return newCursorRequest(req), nil
}

// OpenKeyCursorKey is the same as OpenKeyCursor, but opens a cursor over the given key instead.
func (o *ObjectStore) OpenKeyCursorKey(key js.Value, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openKeyCursor", key, direction.String()))
	return newCursorRequest(req), nil
}

// OpenKeyCursorRange is the same as OpenKeyCursor, but opens a cursor over the given key range instead.
func (o *ObjectStore) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openKeyCursor", keyRange, direction.String()))
	return newCursorRequest(req), nil
}
