// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

// IndexOptions contains all options used to create an Index
type IndexOptions struct {
	// Unique disallows duplicate values for a single key.
	Unique bool
	// MultiEntry adds an entry in the index for each array element when the keyPath resolves to an Array. If false, adds one single entry containing the Array.
	MultiEntry bool
}

// Index provides asynchronous access to an index in a database. An index is a kind of object store for looking up records in another object store, called the referenced object store. You use this to retrieve data.
type Index struct {
	base *baseObjectStore // don't embed to avoid generated docs with the wrong receiver type (Index vs *Index)
}

func wrapIndex(txn *Transaction, jsIndex js.Value) *Index {
	return &Index{wrapBaseObjectStore(txn, jsIndex)}
}

// ObjectStore returns the object store referenced by this index.
func (i *Index) ObjectStore() (_ *ObjectStore, err error) {
	defer exception.Catch(&err)
	return wrapObjectStore(i.base.txn, i.base.jsObjectStore.Get("objectStore")), nil
}

// Name returns the name of this index
func (i *Index) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return i.base.jsObjectStore.Get("name").String(), nil
}

// KeyPath returns the key path of this index. If js.Null(), this index is not auto-populated.
func (i *Index) KeyPath() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return i.base.jsObjectStore.Get("keyPath"), nil
}

// MultiEntry affects how the index behaves when the result of evaluating the index's key path yields an array. If true, there is one record in the index for each item in an array of keys. If false, then there is one record for each key that is an array.
func (i *Index) MultiEntry() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.base.jsObjectStore.Get("multiEntry").Bool(), nil
}

// Unique indicates this index does not allow duplicate values for a key.
func (i *Index) Unique() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.base.jsObjectStore.Get("unique").Bool(), nil
}

// Count returns a UintRequest, and, in a separate thread, returns the total number of records in the index.
func (i *Index) Count() (*UintRequest, error) {
	return i.base.Count()
}

// CountKey returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided key.
func (i *Index) CountKey(key js.Value) (*UintRequest, error) {
	return i.base.CountKey(key)
}

// CountRange returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided KeyRange.
func (i *Index) CountRange(keyRange *KeyRange) (*UintRequest, error) {
	return i.base.CountRange(keyRange)
}

// GetAllKeys returns an ArrayRequest that retrieves record keys for all objects in the index.
func (i *Index) GetAllKeys() (*ArrayRequest, error) {
	return i.base.GetAllKeys()
}

// GetAllKeysRange returns an ArrayRequest that retrieves record keys for all objects in the index matching the specified query. If maxCount is 0, retrieves all objects matching the query.
func (i *Index) GetAllKeysRange(query *KeyRange, maxCount uint) (*ArrayRequest, error) {
	return i.base.GetAllKeysRange(query, maxCount)
}

// Get returns a Request, and, in a separate thread, returns objects selected by the specified key. This is for retrieving specific records from an index.
func (i *Index) Get(key js.Value) (*Request, error) {
	return i.base.Get(key)
}

// GetKey returns a Request, and, in a separate thread retrieves and returns the record key for the object matching the specified parameter.
func (i *Index) GetKey(value js.Value) (*Request, error) {
	return i.base.GetKey(value)
}

// OpenCursor returns a CursorWithValueRequest, and, in a separate thread, returns a new CursorWithValue. Used for iterating through an index by primary key with a cursor.
func (i *Index) OpenCursor(direction CursorDirection) (*CursorWithValueRequest, error) {
	return i.base.OpenCursor(direction)
}

// OpenCursorKey is the same as OpenCursor, but opens a cursor over the given key instead.
func (i *Index) OpenCursorKey(key js.Value, direction CursorDirection) (*CursorWithValueRequest, error) {
	return i.base.OpenCursorKey(key, direction)
}

// OpenCursorRange is the same as OpenCursor, but opens a cursor over the given range instead.
func (i *Index) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorWithValueRequest, error) {
	return i.base.OpenCursorRange(keyRange, direction)
}

// OpenKeyCursor returns a CursorRequest, and, in a separate thread, returns a new Cursor. Used for iterating through all keys in an object store.
func (i *Index) OpenKeyCursor(direction CursorDirection) (*CursorRequest, error) {
	return i.base.OpenKeyCursor(direction)
}

// OpenKeyCursorKey is the same as OpenKeyCursor, but opens a cursor over the given key instead.
func (i *Index) OpenKeyCursorKey(key js.Value, direction CursorDirection) (*CursorRequest, error) {
	return i.base.OpenKeyCursorKey(key, direction)
}

// OpenKeyCursorRange is the same as OpenKeyCursor, but opens a cursor over the given key range instead.
func (i *Index) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (*CursorRequest, error) {
	return i.base.OpenKeyCursorRange(keyRange, direction)
}
