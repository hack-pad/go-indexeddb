//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
)

func TestTransactionDatabase(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)

	txnDB, err := txn.Database()
	assert.NoError(t, err)
	assert.Equal(t, db.jsDB, txnDB.jsDB)
}

func TestTransactionDurability(t *testing.T) {
	t.Parallel()
	const storeName = "mystore"
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore(storeName, ObjectStoreOptions{})
		assert.NoError(t, err)
	})

	for _, tc := range []struct {
		durability   TransactionDurability
		expectString string
	}{
		{
			durability:   DurabilityDefault,
			expectString: "default",
		},
		{
			durability:   DurabilityRelaxed,
			expectString: "relaxed",
		},
		{
			durability:   DurabilityStrict,
			expectString: "strict",
		},
	} {
		tc := tc // enable parallel sub-tests
		t.Run(tc.durability.String(), func(t *testing.T) {
			t.Parallel()
			txn, err := db.TransactionWithOptions(TransactionOptions{
				Durability: tc.durability,
			}, storeName)
			assert.NoError(t, err)
			dur, err := txn.Durability()
			assert.NoError(t, err)
			assert.Equal(t, tc.durability, dur)
			assert.Equal(t, tc.expectString, dur.String())
		})
	}
}

func TestTransactionAbortErr(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadWrite, "mystore")
	assert.NoError(t, err)
	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	_, err = store.AddKey(js.ValueOf("some id"), js.ValueOf(nil))
	assert.NoError(t, err)

	resultErr := txn.listenFinished(context.Background())
	assert.NoError(t, txn.Abort())
	err = <-resultErr
	assert.ErrorIs(t, ErrAborted, err)
	err = txn.Err()
	assert.NoError(t, err)
}

func TestTransactionMode(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})

	t.Run("read only", func(t *testing.T) {
		t.Parallel()
		txn, err := db.Transaction(TransactionReadOnly, "mystore")
		assert.NoError(t, err)

		mode, err := txn.Mode()
		assert.NoError(t, err)
		assert.Equal(t, TransactionReadOnly, mode)
	})

	t.Run("read write", func(t *testing.T) {
		t.Parallel()
		txn, err := db.Transaction(TransactionReadWrite, "mystore")
		assert.NoError(t, err)

		mode, err := txn.Mode()
		assert.NoError(t, err)
		assert.Equal(t, TransactionReadWrite, mode)
	})
}

func TestTransactionObjectStoreNames(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)

	names, err := txn.ObjectStoreNames()
	assert.NoError(t, err)
	assert.Equal(t, []string{"mystore"}, names)
}

func TestTransactionObjectStore(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)

	store, err := txn.ObjectStore("mystore")
	assert.NoError(t, err)
	assert.NotZero(t, store)

	_, err = txn.ObjectStore("not a store")
	assert.Error(t, err)
}

func TestTransactionCommit(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	txn, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)

	err = txn.Commit()
	assert.NoError(t, err)

	assert.NoError(t, txn.Await(context.Background()))

	err = txn.Commit()
	assert.Error(t, err)
}
