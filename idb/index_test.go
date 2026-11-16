//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
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

func TestIndexGetAllKeys(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		records      []map[string]interface{}
		getFn        func(*Index) (*ArrayRequest, error)
		expectResult []js.Value
	}{
		{
			name: "get all keys",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
				{"id": "3", "tag": "b"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAllKeys()
			},
			expectResult: []js.Value{js.ValueOf("2"), js.ValueOf("1"), js.ValueOf("3")},
		},
		{
			name: "get all keys query",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
				{"id": "3", "tag": "b"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				keyRange, err := NewKeyRangeOnly(js.ValueOf("b"))
				assert.NoError(t, err)
				return index.GetAllKeysRange(keyRange, 1)
			},
			expectResult: []js.Value{js.ValueOf("1")},
		},
		{
			name: "get all keys nil range with maxCount",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
				{"id": "3", "tag": "c"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAllKeysRange(nil, 2)
			},
			expectResult: []js.Value{js.ValueOf("2"), js.ValueOf("1")},
		},
		{
			name: "get all keys nil range unbounded",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAllKeysRange(nil, 0)
			},
			expectResult: []js.Value{js.ValueOf("2"), js.ValueOf("1")},
		},
		{
			name: "get all keys empty index",
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAllKeys()
			},
			expectResult: []js.Value(nil),
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
					KeyPath: js.ValueOf("id"),
				})
				assert.NoError(t, err)
				_, err = store.CreateIndex("byTag", js.ValueOf("tag"), IndexOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)
			for _, rec := range tc.records {
				_, err = store.Add(js.ValueOf(rec))
				assert.NoError(t, err)
			}
			index, err := store.Index("byTag")
			assert.NoError(t, err)

			req, err := tc.getFn(index)
			assert.NoError(t, err)
			result, err := req.Await(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tc.expectResult, result)
		})
	}
}

func TestIndexGetAll(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		records   []map[string]interface{}
		getFn     func(*Index) (*ArrayRequest, error)
		expectIDs []string
	}{
		{
			name: "get all",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
				{"id": "3", "tag": "b"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAll()
			},
			expectIDs: []string{"2", "1", "3"},
		},
		{
			name: "get all key matching multiple",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
				{"id": "3", "tag": "b"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAllKey(js.ValueOf("b"))
			},
			expectIDs: []string{"1", "3"},
		},
		{
			name: "get all range",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
				{"id": "3", "tag": "b"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				keyRange, err := NewKeyRangeOnly(js.ValueOf("b"))
				assert.NoError(t, err)
				return index.GetAllRange(keyRange, 1)
			},
			expectIDs: []string{"1"},
		},
		{
			name: "get all nil range with maxCount",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
				{"id": "3", "tag": "c"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAllRange(nil, 2)
			},
			expectIDs: []string{"2", "1"},
		},
		{
			name: "get all nil range unbounded",
			records: []map[string]interface{}{
				{"id": "1", "tag": "b"},
				{"id": "2", "tag": "a"},
			},
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAllRange(nil, 0)
			},
			expectIDs: []string{"2", "1"},
		},
		{
			name: "get all empty index",
			getFn: func(index *Index) (*ArrayRequest, error) {
				return index.GetAll()
			},
			expectIDs: nil,
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
					KeyPath: js.ValueOf("id"),
				})
				assert.NoError(t, err)
				_, err = store.CreateIndex("byTag", js.ValueOf("tag"), IndexOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)
			for _, rec := range tc.records {
				_, err = store.Add(js.ValueOf(rec))
				assert.NoError(t, err)
			}
			index, err := store.Index("byTag")
			assert.NoError(t, err)

			req, err := tc.getFn(index)
			assert.NoError(t, err)
			result, err := req.Await(context.Background())
			assert.NoError(t, err)

			var gotIDs []string
			for _, val := range result {
				gotIDs = append(gotIDs, val.Get("id").String())
			}
			assert.Equal(t, tc.expectIDs, gotIDs)
		})
	}
}
