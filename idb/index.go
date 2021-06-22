// +build js,wasm

package idb

import (
	"log"
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

func (i *Index) Count() (_ <-chan int, err error) {
	defer exception.Catch(&err)
	count := make(chan int)
	req := wrapRequest(i.jsIndex.Call("count"))
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
	return count, nil
}

func (i *Index) Get(key js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("get", key)), nil
}

func (i *Index) GetKey(value js.Value) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("getKey", value)), nil
}

func (i *Index) GetAllKeys(query js.Value) (vals *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(i.jsIndex.Call("getAllKeys", query)), nil
}

func (i *Index) OpenCursor(key js.Value, direction CursorDirection) (_ <-chan *Cursor, err error) {
	defer exception.Catch(&err)
	cursor := make(chan *Cursor)
	req := wrapRequest(i.jsIndex.Call("openCursor", key, direction.String()))
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

//func (i *Index) OpenKeyCursor()
