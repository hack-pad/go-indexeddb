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
	jsIndex js.Value
}

func wrapIndex(jsIndex js.Value) *Index {
	return &Index{
		jsIndex: jsIndex,
	}
}

// Name returns the name of this index
func (i *Index) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("name").String(), nil
}

// ObjectStore returns the name of the object store referenced by this index.
func (i *Index) ObjectStore() (_ *ObjectStore, err error) {
	defer exception.Catch(&err)
	return wrapObjectStore(i.jsIndex.Get("objectStore")), nil
}

// KeyPath returns the key path of this index. If js.Null(), this index is not auto-populated.
func (i *Index) KeyPath() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("keyPath"), nil
}

// MultiEntry affects how the index behaves when the result of evaluating the index's key path yields an array. If true, there is one record in the index for each item in an array of keys. If false, then there is one record for each key that is an array.
func (i *Index) MultiEntry() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("multiEntry").Bool(), nil
}

// Unique indicates this index does not allow duplicate values for a key.
func (i *Index) Unique() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("unique").Bool(), nil
}

// Count returns a UintRequest and returns the number of records within a key range.
func (i *Index) Count() (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("count"))
	return newUintRequest(req), nil
}

// CountKey returns a UintRequest and returns the number of records within a key range.
func (i *Index) CountKey(key js.Value) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("count", key))
	return newUintRequest(req), nil
}

// CountRange returns a UintRequest and returns the number of records within a key range.
func (i *Index) CountRange(keyRange *KeyRange) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("count", keyRange))
	return newUintRequest(req), nil
}

// Get returns a Request and finds the value in the referenced object store that corresponds to the given key.
func (i *Index) Get(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("get", key)), nil
}

// GetRange returns a Request and finds the first corresponding value in the given KeyRange.
func (i *Index) GetRange(keyRange *KeyRange) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("get", keyRange)), nil
}

// GetKey returns a Request and finds the given key.
func (i *Index) GetKey(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("getKey", key)), nil
}

// GetKeyInRange returns a Request and finds the given primary key.
func (i *Index) GetKeyInRange(keyRange *KeyRange) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("getKey", keyRange)), nil
}

// GetAllKeys returns an ArrayRequest, finds all matching keys in the referenced object store that correspond to the given key.
func (i *Index) GetAllKeys(query js.Value) (_ *ArrayRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("getAllKeys", query))
	return newArrayRequest(req), nil
}

// GetAllKeysRange returns an ArrayRequest, finds all matching keys in the referenced object store that are in range.
func (i *Index) GetAllKeysRange(keyRange *KeyRange) (_ *ArrayRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("getAllKeys", keyRange))
	return newArrayRequest(req), nil
}

// OpenCursor returns a CursorWithValueRequest and creates a cursor over the specified key.
func (i *Index) OpenCursor(key js.Value, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("openCursor", key, direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenCursorRange returns a CursorWithValueRequest and creates a cursor over the specified key range.
func (i *Index) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("openCursor", keyRange, direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenCursorAll returns a CursorWithValueRequest and creates a cursor over all keys.
func (i *Index) OpenCursorAll(direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("openCursor", nil, direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenKeyCursor returns a CursorRequest and creates a cursor over the specified key, as arranged by this index.
func (i *Index) OpenKeyCursor(key js.Value, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("openKeyCursor", key, direction.String()))
	return newCursorRequest(req), nil
}

// OpenKeyCursorRange returns a CursorRequest and creates a cursor over the specified key range, as arranged by this index.
func (i *Index) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("openKeyCursor", keyRange, direction.String()))
	return newCursorRequest(req), nil
}

// OpenKeyCursorAll returns a CursorRequest and creates a cursor over all keys, as arranged by this index.
func (i *Index) OpenKeyCursorAll(direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(i.jsIndex.Call("openKeyCursor", nil, direction.String()))
	return newCursorRequest(req), nil
}
