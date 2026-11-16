//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"sync/atomic"
	"syscall/js"
	"testing"
	"time"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
)

var (
	testRequestKey = js.ValueOf("key")
)

func testRequest(t *testing.T) (*Transaction, *Request) {
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	store, err := txn.ObjectStore("mystore")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	req, err := store.PutKey(testRequestKey, js.ValueOf("value"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return txn, req
}

func TestRequestSource(t *testing.T) {
	t.Parallel()
	_, req := testRequest(t)
	store, index, err := req.Source()
	assert.NoError(t, err)
	assert.NotZero(t, store)
	assert.Zero(t, index)
}

func TestRequestAwait(t *testing.T) {
	t.Parallel()
	_, req := testRequest(t)

	result, err := req.Await(context.Background())
	assert.Equal(t, testRequestKey, result)
	assert.NoError(t, err)

	result, err = req.Result()
	assert.NoError(t, err)
	assert.Equal(t, testRequestKey, result)

	err = req.Err()
	assert.NoError(t, err)
}

func TestRequestReadyState(t *testing.T) {
	t.Parallel()
	_, req := testRequest(t)

	result, err := req.Await(context.Background())
	assert.Equal(t, testRequestKey, result)
	assert.NoError(t, err)

	state, err := req.ReadyState()
	assert.NoError(t, err)
	assert.Equal(t, "done", state)
}

func TestRequestTransaction(t *testing.T) {
	t.Parallel()
	txn, req := testRequest(t)

	reqTxn, err := req.Transaction()
	assert.NoError(t, err)
	assert.Equal(t, txn.jsTransaction, reqTxn.jsTransaction)
}

func TestListen(t *testing.T) {
	t.Parallel()
	_, req := testRequest(t)

	var successCount int64
	err := req.Listen(context.Background(), func() {
		atomic.AddInt64(&successCount, 1)
		result, err := req.Result()
		assert.NoError(t, err)
		assert.Equal(t, testRequestKey, result)
	}, func() {
		t.Error("Failed should not be called:", req.Err())
	})
	assert.NoError(t, err)

	assert.Eventually(t, func(ctx context.Context) bool {
		return atomic.LoadInt64(&successCount) > 0
	}, time.Second, 50*time.Millisecond)
}

func TestArrayRequestResult(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		keys         []string
		expectResult []js.Value
	}{
		{
			name:         "empty array",
			keys:         nil,
			expectResult: []js.Value(nil),
		},
		{
			name:         "populated array",
			keys:         []string{"alpha", "beta", "gamma"},
			expectResult: []js.Value{js.ValueOf("alpha"), js.ValueOf("beta"), js.ValueOf("gamma")},
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
			for _, key := range tc.keys {
				_, err := store.AddKey(js.ValueOf(key), js.ValueOf(key))
				assert.NoError(t, err)
			}

			req, err := store.GetAllKeys()
			assert.NoError(t, err)

			awaited, err := req.Await(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tc.expectResult, awaited)

			result, err := req.Result()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectResult, result)
		})
	}
}
