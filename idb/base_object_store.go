// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

// baseObjectStore is the common implementation for both object stores and indexes.
type baseObjectStore struct {
	jsObjectStore js.Value
}

func wrapBaseObjectStore(jsObjectStore js.Value) *baseObjectStore {
	return &baseObjectStore{jsObjectStore: jsObjectStore}
}

// Count returns a UintRequest, and, in a separate thread, returns the total number of records in the store or index.
func (b *baseObjectStore) Count() (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("count"))
	return newUintRequest(req), nil
}

// CountKey returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided key.
func (b *baseObjectStore) CountKey(key js.Value) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("count", key))
	return newUintRequest(req), nil
}

// CountRange returns a UintRequest, and, in a separate thread, returns the total number of records that match the provided KeyRange.
func (b *baseObjectStore) CountRange(keyRange *KeyRange) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("count", keyRange))
	return newUintRequest(req), nil
}

// GetAllKeys returns an ArrayRequest that retrieves record keys for all objects in the object store or index.
func (b *baseObjectStore) GetAllKeys() (_ *ArrayRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("getAllKeys"))
	return newArrayRequest(req), nil
}

// GetAllKeysRange returns an ArrayRequest that retrieves record keys for all objects in the object store or index matching the specified query. If maxCount is 0, retrieves all objects matching the query.
func (b *baseObjectStore) GetAllKeysRange(query *KeyRange, maxCount uint) (_ *ArrayRequest, err error) {
	defer exception.Catch(&err)
	args := []interface{}{query}
	if maxCount > 0 {
		args = append(args, maxCount)
	}
	req := wrapRequest(b.jsObjectStore.Call("getAllKeys", args...))
	return newArrayRequest(req), nil
}

// Get returns a Request, and, in a separate thread, returns the objects selected by the specified key. This is for retrieving specific records from an object store or index.
func (b *baseObjectStore) Get(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(b.jsObjectStore.Call("get", key)), nil
}

// GetKey returns a Request, and, in a separate thread retrieves and returns the record key for the object matching the specified parameter.
func (b *baseObjectStore) GetKey(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(b.jsObjectStore.Call("getKey", value)), nil
}

// OpenCursor returns a CursorWithValueRequest, and, in a separate thread, returns a new CursorWithValue. Used for iterating through an object store or index by primary key with a cursor.
func (b *baseObjectStore) OpenCursor(direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("openCursor", js.Null(), direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenCursorKey is the same as OpenCursor, but opens a cursor over the given key instead.
func (b *baseObjectStore) OpenCursorKey(key js.Value, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("openCursor", key, direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenCursorRange is the same as OpenCursor, but opens a cursor over the given range instead.
func (b *baseObjectStore) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("openCursor", keyRange, direction.String()))
	return newCursorWithValueRequest(req), nil
}

// OpenKeyCursor returns a CursorRequest, and, in a separate thread, returns a new Cursor. Used for iterating through all keys in an object store or index.
func (b *baseObjectStore) OpenKeyCursor(direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("openKeyCursor", js.Null(), direction.String()))
	return newCursorRequest(req), nil
}

// OpenKeyCursorKey is the same as OpenKeyCursor, but opens a cursor over the given key instead.
func (b *baseObjectStore) OpenKeyCursorKey(key js.Value, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("openKeyCursor", key, direction.String()))
	return newCursorRequest(req), nil
}

// OpenKeyCursorRange is the same as OpenKeyCursor, but opens a cursor over the given key range instead.
func (b *baseObjectStore) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(b.jsObjectStore.Call("openKeyCursor", keyRange, direction.String()))
	return newCursorRequest(req), nil
}
