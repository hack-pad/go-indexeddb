// +build js,wasm

package idb

import (
	"sync"
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
	"github.com/pkg/errors"
)

type Factory struct {
	jsFactory js.Value
}

var (
	global     *Factory
	globalErr  error
	globalOnce sync.Once
)

// Global returns the global IndexedDB instance.
// Can be called multiple times, will always return the same result (or error if one occurs).
func Global() (*Factory, error) {
	globalOnce.Do(func() {
		jsFactory := js.Global().Get("indexedDB")
		if !jsFactory.Truthy() {
			globalErr = errors.New("Global JS variable 'indexedDB' is not defined.")
		} else {
			global, globalErr = WrapFactory(jsFactory)
		}
	})
	return global, globalErr
}

func WrapFactory(jsFactory js.Value) (*Factory, error) {
	return &Factory{jsFactory: jsFactory}, nil
}

// Open requests to open a connection to a database.
func (f *Factory) Open(name string, version uint, upgrader Upgrader) (_ *OpenDBRequest, err error) {
	defer exception.Catch(&err)

	args := []interface{}{name}
	if version > 0 {
		args = append(args, version)
	}
	req := wrapRequest(f.jsFactory.Call("open", args...))
	return wrapOpenDBRequest(req, upgrader), nil
}

// DeleteDatabase requests the deletion of a database.
func (f *Factory) DeleteDatabase(name string) (_ *Request, err error) {
	defer exception.Catch(&err)
	return wrapRequest(f.jsFactory.Call("deleteDatabase", name)), nil
}

// CompareKeys compares two keys and returns a result indicating which one is greater in value.
func (f *Factory) CompareKeys(a, b js.Value) (_ int, err error) {
	defer exception.Catch(&err)
	compare := f.jsFactory.Call("cmp", a, b)
	return compare.Int(), nil
}
