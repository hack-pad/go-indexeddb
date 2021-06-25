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
	// ObjectStore is the object store interface for this index.
	// Do not use directly. To access the underlying ObjectStore, use GetObjectStore().
	*ObjectStore
}

func wrapIndex(jsIndex js.Value) *Index {
	return &Index{
		ObjectStore: wrapObjectStore(jsIndex),
	}
}

// GetObjectStore returns the object store referenced by this index.
func (i *Index) GetObjectStore() (_ *ObjectStore, err error) {
	defer exception.Catch(&err)
	return wrapObjectStore(i.jsObjectStore.Get("objectStore")), nil
}

// Name returns the name of this index
func (i *Index) Name() (_ string, err error) {
	defer exception.Catch(&err)
	return i.jsObjectStore.Get("name").String(), nil
}

// KeyPath returns the key path of this index. If js.Null(), this index is not auto-populated.
func (i *Index) KeyPath() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return i.jsObjectStore.Get("keyPath"), nil
}

// MultiEntry affects how the index behaves when the result of evaluating the index's key path yields an array. If true, there is one record in the index for each item in an array of keys. If false, then there is one record for each key that is an array.
func (i *Index) MultiEntry() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.jsObjectStore.Get("multiEntry").Bool(), nil
}

// Unique indicates this index does not allow duplicate values for a key.
func (i *Index) Unique() (_ bool, err error) {
	defer exception.Catch(&err)
	return i.jsObjectStore.Get("unique").Bool(), nil
}
