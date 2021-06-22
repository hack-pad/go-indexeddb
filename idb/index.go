// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

type IndexOptions struct {
	Unique     bool
	MultiEntry bool
}

type Index struct {
	jsIndex js.Value
}

func wrapIndex(jsIndex js.Value) *Index {
	return &Index{
		jsIndex: jsIndex,
	}
}

func (i *Index) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("name").String(), nil
}

func (i *Index) ObjectStore() (_ *ObjectStore, err error) {
	defer exception.Catch(&err)
	return wrapObjectStore(i.jsIndex.Get("objectStore")), nil
}

func (i *Index) KeyPath() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("keyPath"), nil
}

func (i *Index) MultiEntry() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("multiEntry").Bool(), nil
}

func (i *Index) Unique() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.jsIndex.Get("unique").Bool(), nil
}

func (i *Index) Count() (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("count")), nil
}

func (i *Index) CountKey(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("count", key)), nil
}

func (i *Index) CountRange(keyRange *KeyRange) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("count", keyRange)), nil
}

func (i *Index) Get(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("get", key)), nil
}

func (i *Index) GetKey(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("getKey", value)), nil
}

func (i *Index) GetAllKeys(query js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("getAllKeys", query)), nil
}

func (i *Index) OpenCursor(key js.Value, direction CursorDirection) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("openCursor", key, direction.String())), nil
}

func (i *Index) OpenCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("openCursor", keyRange, direction.String())), nil
}

func (i *Index) OpenCursorAll(direction CursorDirection) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("openCursor", nil, direction.String())), nil
}

func (i *Index) OpenKeyCursor(key js.Value, direction CursorDirection) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("openKeyCursor", key, direction.String())), nil
}

func (i *Index) OpenKeyCursorRange(keyRange *KeyRange, direction CursorDirection) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("openKeyCursor", keyRange, direction.String())), nil
}

func (i *Index) OpenKeyCursorAll(direction CursorDirection) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("openKeyCursor", nil, direction.String())), nil
}
