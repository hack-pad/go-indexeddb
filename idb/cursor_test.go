// +build js,wasm

package idb

import (
	"context"
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
)

var (
	someKeyStoreData = [][]interface{}{
		{"some id 1", map[string]interface{}{"primary": "some value 1"}},
		{"some id 2", map[string]interface{}{"primary": "some value 2"}},
		{"some id 3", map[string]interface{}{"primary": "some value 3"}},
		{"some id 4", map[string]interface{}{"primary": "some value 4"}},
		{"some id 5", map[string]interface{}{"primary": "some value 5"}},
	}
)

func someKeyStore(tb testing.TB) (*ObjectStore, *Index) {
	tb.Helper()
	db := testDB(tb, func(db *Database) {
		store, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(tb, err)
		_, err = store.CreateIndex("myindex", js.ValueOf("primary"), IndexOptions{})
		assert.NoError(tb, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(tb, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(tb, err)
	index, err := store.Index("myindex")
	assert.NoError(tb, err)

	for _, object := range someKeyStoreData {
		key, value := object[0], object[1]
		_, err := store.AddKey(js.ValueOf(key), js.ValueOf(value))
		assert.NoError(tb, err)
	}
	return store, index
}

func TestCursorSource(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	cursor, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	cursorStore, cursorIndex, err := cursor.Source()
	assert.NoError(t, err)
	assert.Zero(t, cursorIndex)
	assert.Equal(t, store, cursorStore)
}

func TestCursorDirection(t *testing.T) {
	t.Parallel()
	for _, direction := range []CursorDirection{
		CursorNext,
		CursorNextUnique,
		CursorPrevious,
		CursorPreviousUnique,
	} {
		t.Log("Direction:", direction) // disabled parallel subtests here, due to an issue in paralleltest linter
		store, _ := someKeyStore(t)

		req, err := store.OpenCursor(direction)
		assert.NoError(t, err)
		cursor, err := req.Await(context.Background())
		assert.NoError(t, err)

		actualDirection, err := cursor.Direction()
		assert.NoError(t, err)
		assert.Equal(t, direction, actualDirection)
	}
}

func TestCursorKey(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := someKeyStoreData[iterIndex][0]
		key, err := cursor.Key()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData), iterIndex)
}

func TestCursorPrimaryKey(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := someKeyStoreData[iterIndex][0]
		key, err := cursor.PrimaryKey()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData), iterIndex)
}

func TestCursorRequest(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	cursor, err := req.Await(context.Background())
	assert.NoError(t, err)
	cursorReq, err := cursor.Request()
	assert.NoError(t, err)
	assert.Equal(t, req.jsRequest, cursorReq.jsRequest)
}

func TestCursorAdvance(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := someKeyStoreData[iterIndex*2][0]
		key, err := cursor.Key()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)
		err = cursor.Advance(2)
		assert.NoError(t, err)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData)/2+len(someKeyStoreData)%2, iterIndex)
}

func TestCursorContinue(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := someKeyStoreData[iterIndex][0]
		key, err := cursor.Key()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)
		err = cursor.Continue()
		assert.NoError(t, err)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData), iterIndex)
}

func TestCursorContinueKey(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := someKeyStoreData[iterIndex*2][0]
		key, err := cursor.Key()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)

		nextIndex := (iterIndex + 1) * 2
		if nextIndex >= len(someKeyStoreData) {
			return ErrCursorStopIter
		}
		nextKey := someKeyStoreData[nextIndex][0]
		err = cursor.ContinueKey(js.ValueOf(nextKey))
		assert.NoError(t, err)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData)/2, iterIndex)
}

func TestCursorContinuePrimaryKey(t *testing.T) {
	t.Parallel()
	_, index := someKeyStore(t)
	req, err := index.OpenCursor(CursorNext)
	assert.NoError(t, err)

	getPrimaryKey := func(ix int) string {
		return someKeyStoreData[ix][1].(map[string]interface{})["primary"].(string)
	}

	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := getPrimaryKey(iterIndex * 2)
		key, err := cursor.Key()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)

		nextIndex := (iterIndex + 1) * 2
		if nextIndex >= len(someKeyStoreData) {
			return ErrCursorStopIter
		}
		nextKey := someKeyStoreData[nextIndex][0]
		nextIndexKey := getPrimaryKey(nextIndex)
		err = cursor.ContinuePrimaryKey(js.ValueOf(nextIndexKey), js.ValueOf(nextKey))
		assert.NoError(t, err)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData)/2, iterIndex)
}

func TestCursorDelete(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := someKeyStoreData[iterIndex][0]
		key, err := cursor.Key()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)
		_, err = cursor.Delete()
		assert.NoError(t, err)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData), iterIndex)

	// ensure empty after deleting everything
	txn, err := store.Transaction()
	assert.NoError(t, err)
	storeName, err := store.Name()
	assert.NoError(t, err)
	emptyStore, err := txn.ObjectStore(storeName)
	assert.NoError(t, err)
	countReq, err := emptyStore.Count()
	assert.NoError(t, err)
	count, err := countReq.Await(context.Background())
	assert.NoError(t, err)
	assert.Zero(t, count)
}

func TestCursorUpdateAndValue(t *testing.T) {
	t.Parallel()
	store, _ := someKeyStore(t)
	req, err := store.OpenCursor(CursorNext)
	assert.NoError(t, err)

	// set all keys to their iteration index
	iterIndex := 0
	assert.NoError(t, req.Iter(context.Background(), func(cursor *CursorWithValue) error {
		expectKey := someKeyStoreData[iterIndex][0]
		key, err := cursor.Key()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(expectKey), key)
		_, err = cursor.Update(js.ValueOf(iterIndex))
		assert.NoError(t, err)
		iterIndex++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData), iterIndex)

	// ensure values equal indices after updating everything
	txn, err := store.Transaction()
	assert.NoError(t, err)
	storeName, err := store.Name()
	assert.NoError(t, err)
	updatedStore, err := txn.ObjectStore(storeName)
	assert.NoError(t, err)
	cursorReq, err := updatedStore.OpenCursor(CursorNext)
	assert.NoError(t, err)
	ix := 0
	assert.NoError(t, cursorReq.Iter(context.Background(), func(cursor *CursorWithValue) error {
		value, err := cursor.Value()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf(ix), value)
		ix++
		return err
	}))
	assert.Equal(t, len(someKeyStoreData), ix)
}
