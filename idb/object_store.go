// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

type ObjectStoreOptions struct {
	KeyPath       string
	AutoIncrement bool
}

type ObjectStore struct {
	jsObjectStore js.Value
}

func wrapObjectStore(jsObjectStore js.Value) *ObjectStore {
	return &ObjectStore{jsObjectStore: jsObjectStore}
}

func (o *ObjectStore) IndexNames() (_ []string, err error) {
	defer exception.Catch(&err)
	return stringsFromArray(o.jsObjectStore.Get("indexNames"))
}

func (o *ObjectStore) KeyPath() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return o.jsObjectStore.Get("keyPath"), nil
}

func (o *ObjectStore) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return o.jsObjectStore.Get("name").String(), nil
}

func (o *ObjectStore) Transaction() (_ *Transaction, err error) {
	defer exception.Catch(&err)
	return wrapTransaction(o.jsObjectStore.Get("transaction")), nil
}

func (o *ObjectStore) AutoIncrement() (_ bool, err error) {
	defer exception.Catch(&err)
	return o.jsObjectStore.Get("autoIncrement").Bool(), nil
}

func (o *ObjectStore) Add(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("add", value)), nil
}

func (o *ObjectStore) AddKey(key, value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("add", value, key)), nil
}

func (o *ObjectStore) Clear() (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("clear"))
	return newAckRequest(req), nil
}

func (o *ObjectStore) Count() (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("count"))
	return newUintRequest(req), nil
}

func (o *ObjectStore) CountKey(key js.Value) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("count", key))
	return newUintRequest(req), nil
}

func (o *ObjectStore) CountRange(keyRange *KeyRange) (_ *UintRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("count", keyRange))
	return newUintRequest(req), nil
}

func (o *ObjectStore) CreateIndex(name string, keyPath js.Value, options IndexOptions) (index *Index, err error) {
	defer exception.Catch(&err)
	jsIndex := o.jsObjectStore.Call("createIndex", name, keyPath, map[string]interface{}{
		"unique":     options.Unique,
		"multiEntry": options.MultiEntry,
	})
	return wrapIndex(jsIndex), nil
}

func (o *ObjectStore) Delete(key js.Value) (_ *AckRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("delete", key))
	return newAckRequest(req), nil
}

func (o *ObjectStore) DeleteIndex(name string) (err error) {
	defer exception.Catch(&err)
	o.jsObjectStore.Call("deleteIndex", name)
	return nil
}

func (o *ObjectStore) Get(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("get", key)), nil
}

func (o *ObjectStore) GetAllKeys(query js.Value) (_ *ArrayRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("getAllKeys", query))
	return newArrayRequest(req), nil
}

func (o *ObjectStore) GetKey(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("getKey", value)), nil
}

func (o *ObjectStore) Index(name string) (index *Index, err error) {
	defer exception.Catch(&err)
	jsIndex := o.jsObjectStore.Call("index", name)
	return wrapIndex(jsIndex), nil
}

func (o *ObjectStore) Put(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("put", value)), nil
}

func (o *ObjectStore) PutKey(key, value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("put", value, key)), nil
}

func (o *ObjectStore) OpenCursor(key js.Value, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openCursor", key, direction.String()))
	return newCursorWithValueRequest(req), nil
}

func (o *ObjectStore) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openCursor", keyRange, direction.String()))
	return newCursorWithValueRequest(req), nil
}

func (o *ObjectStore) OpenCursorAll(direction CursorDirection) (_ *CursorWithValueRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openCursor", nil, direction.String()))
	return newCursorWithValueRequest(req), nil
}

func (o *ObjectStore) OpenKeyCursor(key js.Value, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openKeyCursor", key, direction.String()))
	return newCursorRequest(req), nil
}

func (o *ObjectStore) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openKeyCursor", keyRange, direction.String()))
	return newCursorRequest(req), nil
}

func (o *ObjectStore) OpenKeyCursorAll(direction CursorDirection) (_ *CursorRequest, err error) {
	defer exception.Catch(&err)
	req := wrapRequest(o.jsObjectStore.Call("openKeyCursor", nil, direction.String()))
	return newCursorRequest(req), nil
}
