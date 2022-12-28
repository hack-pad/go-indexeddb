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
	req.Listen(context.Background(), func() {
		atomic.AddInt64(&successCount, 1)
		result, err := req.Result()
		assert.NoError(t, err)
		assert.Equal(t, testRequestKey, result)
	}, func() {
		t.Error("Failed should not be called:", req.Err())
	})

	assert.Eventually(t, func(ctx context.Context) bool {
		return atomic.LoadInt64(&successCount) > 0
	}, time.Second, 50*time.Millisecond)
}
