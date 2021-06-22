// +build js,wasm

package idb

import (
	"log"
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

func (o *ObjectStore) Add(key, value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("add", value, key)), nil
}

func (o *ObjectStore) Clear() (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("clear")), nil
}

func (o *ObjectStore) Count() (_ <-chan int, err error) {
	defer exception.Catch(&err)
	count := make(chan int)
	req := wrapRequest(o.jsObjectStore.Call("count"))
	req.Listen(func() {
		result, err := req.Result()
		if err == nil {
			count <- result.Int()
		} else {
			log.Println("Failed to get count result:", err)
		}
		close(count)
	}, func() {
		close(count)
	})
	return count, err
}

func (o *ObjectStore) CreateIndex(name string, keyPath js.Value, options IndexOptions) (index *Index, err error) {
	defer exception.Catch(&err)
	jsIndex := o.jsObjectStore.Call("createIndex", name, keyPath, map[string]interface{}{
		"unique":     options.Unique,
		"multiEntry": options.MultiEntry,
	})
	return wrapIndex(jsIndex), nil
}

func (o *ObjectStore) Delete(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("delete", key)), nil
}

func (o *ObjectStore) DeleteIndex(name string) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("deleteIndex", name)), nil
}

func (o *ObjectStore) Get(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("get", key)), nil
}

func (o *ObjectStore) GetAllKeys(query js.Value) (vals *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("getAllKeys", query)), nil
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

func (o *ObjectStore) OpenCursor(key js.Value, direction CursorDirection) (_ <-chan *Cursor, err error) {
	defer exception.Catch(&err)
	cursor := make(chan *Cursor)
	req := wrapRequest(o.jsObjectStore.Call("openCursor", key, direction.String()))
	req.Listen(func() {
		result, err := req.Result()
		if err == nil {
			cursor <- &Cursor{jsCursor: result}
		} else {
			log.Println("Failed to get cursor result:", err)
		}
		close(cursor)
	}, func() {
		close(cursor)
	})
	return cursor, nil
}

/*
func (o *ObjectStore) OpenKeyCursor(keyRange KeyRange, direction CursorDirection) (*Cursor, error) {
	panic("not implemented")
}
*/

func (o *ObjectStore) Put(key, value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(o.jsObjectStore.Call("put", value, key)), nil
}
