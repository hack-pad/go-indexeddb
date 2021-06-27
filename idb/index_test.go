// +build js,wasm

package idb

import (
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
)

func TestIndexObjectStore(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	index, err := store.Index("myindex")
	assert.NoError(t, err)

	indexStore, err := index.ObjectStore()
	assert.NoError(t, err)
	assert.Equal(t, store, indexStore)
}

func TestIndexName(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	index, err := store.Index("myindex")
	assert.NoError(t, err)

	name, err := index.Name()
	assert.NoError(t, err)
	assert.Equal(t, "myindex", name)
}

func TestIndexKeyPath(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	index, err := store.Index("myindex")
	assert.NoError(t, err)

	keyPath, err := index.KeyPath()
	assert.NoError(t, err)
	assert.Equal(t, js.ValueOf("primary"), keyPath)
}

func TestIndexMultiEntry(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{
			MultiEntry: true,
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	index, err := store.Index("myindex")
	assert.NoError(t, err)

	multiEntry, err := index.MultiEntry()
	assert.NoError(t, err)
	assert.Equal(t, true, multiEntry)
}

func TestIndexUnique(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{
			Unique: true,
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	index, err := store.Index("myindex")
	assert.NoError(t, err)

	unique, err := index.Unique()
	assert.NoError(t, err)
	assert.Equal(t, true, unique)
}
