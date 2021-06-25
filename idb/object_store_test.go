// +build js,wasm

package idb

import (
	"context"
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
)

func TestObjectStoreIndexNames(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("indexKey"), IndexOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	names, err := store.IndexNames()
	assert.NoError(t, err)
	assert.Equal(t, []string{"myindex"}, names)
}

func TestObjectStoreKeyPath(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("primary"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	keyPath, err := store.KeyPath()
	assert.NoError(t, err)
	assert.Equal(t, js.ValueOf("primary"), keyPath)
}

func TestObjectStoreName(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	name, err := store.Name()
	assert.NoError(t, err)
	assert.Equal(t, "mystore", name)
}

func TestObjectStoreAutoIncrement(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			AutoIncrement: true,
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	autoIncrement, err := store.AutoIncrement()
	assert.NoError(t, err)
	assert.Equal(t, true, autoIncrement)
}

func TestObjectStoreTransaction(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("primary"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	txnGet, err := store.Transaction()
	assert.NoError(t, err)
	assert.Equal(t, txn.jsTransaction, txnGet.jsTransaction)
}

func TestObjectStoreAdd(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("id"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	addReq, err := store.Add(js.ValueOf(map[string]interface{}{
		"id": "some id",
	}))
	assert.NoError(t, err)
	getReq, err := store.GetKey(js.ValueOf("some id"))
	assert.NoError(t, err)

	assert.NoError(t, addReq.Await(context.Background()))
	result, err := getReq.Await(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, js.ValueOf("some id"), result)
}

func TestObjectStoreClear(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	{
		txn, err := db.Transaction(TransactionReadWrite, "mystore")
		assert.NoError(t, err)
		store, err := txn.ObjectStore("mystore")
		assert.NoError(t, err)
		_, err = store.AddKey(js.ValueOf("some key"), js.ValueOf("some value"))
		assert.NoError(t, err)
		assert.NoError(t, txn.Await(context.Background()))
	}

	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	clearReq, err := store.Clear()
	assert.NoError(t, err)
	getReq, err := store.GetAllKeys()
	assert.NoError(t, err)

	assert.NoError(t, clearReq.Await(context.Background()))
	result, err := getReq.Await(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []js.Value(nil), result)
}

func TestObjectStoreCount(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		countFn func(*ObjectStore) (*UintRequest, error)
	}{
		{
			name: "count",
			countFn: func(store *ObjectStore) (*UintRequest, error) {
				return store.Count()
			},
		},
		{
			name: "count key",
			countFn: func(store *ObjectStore) (*UintRequest, error) {
				return store.CountKey(js.ValueOf("some key"))
			},
		},
		{
			name: "count range",
			countFn: func(store *ObjectStore) (*UintRequest, error) {
				keyRange, err := NewKeyRangeOnly(js.ValueOf("some key"))
				assert.NoError(t, err)
				return store.CountRange(keyRange)
			},
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)

			_, err = store.AddKey(js.ValueOf("some key"), js.ValueOf("some value"))
			assert.NoError(t, err)

			req, err := tc.countFn(store)
			assert.NoError(t, err)
			count, err := req.Await(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, uint(1), count)
		})
	}
}

func TestObjectStoreCreateIndex(t *testing.T) {
	t.Parallel()
	testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		index, err := store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{
			Unique:     true,
			MultiEntry: true,
		})
		assert.NoError(t, err)

		unique, err := index.Unique()
		assert.NoError(t, err)
		assert.Equal(t, true, unique)
		multiEntry, err := index.MultiEntry()
		assert.NoError(t, err)
		assert.Equal(t, true, multiEntry)
	})
}

func TestObjectStoreDelete(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	_, err = store.AddKey(js.ValueOf("some key"), js.ValueOf("some value"))
	assert.NoError(t, err)

	_, err = store.Delete(js.ValueOf("some key"))
	assert.NoError(t, err)
	req, err := store.GetKey(js.ValueOf("some key"))
	assert.NoError(t, err)
	result, err := req.Await(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, js.Undefined(), result)
}

func TestObjectStoreDeleteIndex(t *testing.T) {
	t.Parallel()
	testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{})
		assert.NoError(t, err)
		err = store.DeleteIndex("myindex")
		assert.NoError(t, err)
		names, err := store.IndexNames()
		assert.NoError(t, err)
		assert.Equal(t, []string(nil), names)
	})
}

func TestObjectStoreGet(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		keys         map[string]interface{}
		getFn        func(*ObjectStore) (interface{}, error)
		expectResult interface{}
	}{
		{
			name: "get all keys",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				return store.GetAllKeys()
			},
			expectResult: []js.Value{js.ValueOf("some id"), js.ValueOf("some other id")},
		},
		{
			name: "get all keys query",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				keyRange, err := NewKeyRangeOnly(js.ValueOf("some id"))
				assert.NoError(t, err)
				return store.GetAllKeysRange(keyRange, 10)
			},
			expectResult: []js.Value{js.ValueOf("some id")},
		},
		{
			name: "get",
			keys: map[string]interface{}{
				"some id": "some value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				return store.Get(js.ValueOf("some id"))
			},
			expectResult: js.ValueOf("some value"),
		},
		{
			name: "get key",
			keys: map[string]interface{}{
				"some id": "some value",
			},
			getFn: func(store *ObjectStore) (interface{}, error) {
				return store.GetKey(js.ValueOf("some id"))
			},
			expectResult: js.ValueOf("some id"),
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)
			for key, value := range tc.keys {
				_, err := store.AddKey(js.ValueOf(key), js.ValueOf(value))
				assert.NoError(t, err)
			}
			req, err := tc.getFn(store)
			assert.NoError(t, err)
			var result interface{}
			switch req := req.(type) {
			case *ArrayRequest:
				result, err = req.Await(context.Background())
			case *Request:
				result, err = req.Await(context.Background())
			default:
				t.Fatalf("Invalid return type: %T", req)
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectResult, result)
		})
	}
}

func TestObjectStoreIndex(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("indexKey"), IndexOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	index, err := store.Index("myindex")
	assert.NoError(t, err)
	assert.NotZero(t, index)
}

func TestObjectStorePut(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{
			KeyPath: js.ValueOf("id"),
		})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	req, err := store.Put(js.ValueOf(map[string]interface{}{
		"id":    "some id",
		"value": "some value",
	}))
	assert.NoError(t, err)
	resultKey, err := req.Await(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, js.ValueOf("some id"), resultKey)
}

func TestObjectStorePutKey(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)

	req, err := store.PutKey(js.ValueOf("some id"), js.ValueOf("some value"))
	assert.NoError(t, err)
	resultKey, err := req.Await(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, js.ValueOf("some id"), resultKey)
}

func TestObjectStoreOpenCursor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		keys          map[string]interface{}
		cursorFn      func(*ObjectStore) (interface{}, error)
		expectResults []js.Value
	}{
		{
			name: "open cursor next",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenCursor(CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some value"),
				js.ValueOf("some other value"),
			},
		},
		{
			name: "open cursor previous",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenCursor(CursorPrevious)
			},
			expectResults: []js.Value{
				js.ValueOf("some other value"),
				js.ValueOf("some value"),
			},
		},
		{
			name: "open cursor over key",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenCursorKey(js.ValueOf("some id"), CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some value"),
			},
		},
		{
			name: "open cursor over key range",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				keyRange, err := NewKeyRangeLowerBound(js.ValueOf("some more"), true)
				assert.NoError(t, err)
				return store.OpenCursorRange(keyRange, CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some other value"),
			},
		},
		{
			name: "open key cursor",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenKeyCursor(CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some id"),
				js.ValueOf("some other id"),
			},
		},
		{
			name: "open key cursor key",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				return store.OpenKeyCursorKey(js.ValueOf("some id"), CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some id"),
			},
		},
		{
			name: "open key cursor range",
			keys: map[string]interface{}{
				"some id":       "some value",
				"some other id": "some other value",
			},
			cursorFn: func(store *ObjectStore) (interface{}, error) {
				keyRange, err := NewKeyRangeLowerBound(js.ValueOf("some more"), true)
				assert.NoError(t, err)
				return store.OpenKeyCursorRange(keyRange, CursorNext)
			},
			expectResults: []js.Value{
				js.ValueOf("some other id"),
			},
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
				assert.NoError(t, err)
			})
			txn, err := db.Transaction(TransactionReadWrite, "mystore")
			assert.NoError(t, err)
			store, err := txn.ObjectStore("mystore")
			assert.NoError(t, err)
			for key, value := range tc.keys {
				_, err = store.AddKey(js.ValueOf(key), js.ValueOf(value))
				assert.NoError(t, err)
			}

			req, err := tc.cursorFn(store)
			assert.NoError(t, err)
			var results []js.Value
			switch req := req.(type) {
			case *CursorWithValueRequest:
				err := req.Iter(context.Background(), func(cursor *CursorWithValue) error {
					value, err := cursor.Value()
					if assert.NoError(t, err) {
						results = append(results, value)
						err = cursor.Continue()
						return err
					}
					return err
				})
				assert.NoError(t, err)
			case *CursorRequest:
				err := req.Iter(context.Background(), func(cursor *Cursor) error {
					key, err := cursor.Key()
					if assert.NoError(t, err) {
						results = append(results, key)
						return cursor.Continue()
					}
					return err
				})
				assert.NoError(t, err)
			default:
				t.Fatalf("Invalid cursor type: %T", req)
			}
			assert.Equal(t, tc.expectResults, results)
		})
	}
}
