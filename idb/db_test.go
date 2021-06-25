// +build js,wasm

package idb

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
)

func testDB(tb testing.TB, initFunc func(*Database)) *Database {
	tb.Helper()
	dbFactory, err := Global()
	if !assert.NoError(tb, err) {
		tb.FailNow()
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	assert.NoError(tb, err)
	name := fmt.Sprintf("%s%s/%d", testDBPrefix, tb.Name(), n.Int64())
	req, err := dbFactory.Open(context.Background(), name, 0, func(db *Database, oldVersion, newVersion uint) error {
		initFunc(db)
		return nil
	})
	if !assert.NoError(tb, err) {
		tb.FailNow()
	}
	db, err := req.Await(context.Background())
	if !assert.NoError(tb, err) {
		tb.FailNow()
	}
	tb.Cleanup(func() {
		assert.NoError(tb, db.Close())
		req, err := dbFactory.DeleteDatabase(name)
		assert.NoError(tb, err)
		assert.NoError(tb, req.Await(context.Background()))
	})
	return db
}

func TestDatabaseName(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {})
	name, err := db.Name()
	assert.NoError(t, err)
	assert.Contains(t, name, t.Name())
}

func TestDatabaseVersion(t *testing.T) {
	t.Parallel()
	db := testDB(t, func(db *Database) {})
	version, err := db.Version()
	assert.NoError(t, err)
	assert.Equal(t, uint(1), version)
}

func TestDatabaseCreateObjectStore(t *testing.T) {
	t.Parallel()

	t.Run("default options", func(t *testing.T) {
		t.Parallel()
		db := testDB(t, func(db *Database) {
			_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
			assert.NoError(t, err)
		})
		names, err := db.ObjectStoreNames()
		assert.NoError(t, err)
		assert.Equal(t, []string{"mystore"}, names)
	})

	t.Run("set keypath and auto-increment", func(t *testing.T) {
		t.Parallel()
		const storeName = "mystore"
		db := testDB(t, func(db *Database) {
			_, err := db.CreateObjectStore(storeName, ObjectStoreOptions{
				KeyPath:       js.ValueOf("primary"),
				AutoIncrement: true,
			})
			assert.NoError(t, err)
		})
		txn, err := db.Transaction(TransactionReadOnly, storeName)
		assert.NoError(t, err)
		store, err := txn.ObjectStore(storeName)
		assert.NoError(t, err)

		keyPath, err := store.KeyPath()
		assert.NoError(t, err)
		assert.Equal(t, js.ValueOf("primary"), keyPath)
		autoIncrement, err := store.AutoIncrement()
		assert.NoError(t, err)
		assert.Equal(t, true, autoIncrement)
	})
}

func TestDatabaseDeleteObjectStore(t *testing.T) {
	t.Parallel()

	t.Run("delete object store", func(t *testing.T) {
		t.Parallel()
		testDB(t, func(db *Database) {
			_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
			assert.NoError(t, err)

			assert.NoError(t, db.DeleteObjectStore("mystore"))
			names, err := db.ObjectStoreNames()
			assert.NoError(t, err)
			assert.Equal(t, []string(nil), names)
		})
	})

	t.Run("not upgrading", func(t *testing.T) {
		t.Parallel()
		db := testDB(t, func(db *Database) {})
		assert.Error(t, db.DeleteObjectStore("mystore"))
	})
}

func TestDatabaseTransaction(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		makeTxn func(*Database) (*Transaction, error)
	}{
		{
			name: "simple readwrite",
			makeTxn: func(db *Database) (*Transaction, error) {
				return db.Transaction(TransactionReadWrite, "store1", "store2")
			},
		},
		{
			name: "readwrite with durability",
			makeTxn: func(db *Database) (*Transaction, error) {
				return db.TransactionWithOptions(TransactionOptions{
					Mode:       TransactionReadWrite,
					Durability: DurabilityRelaxed,
				}, "store1", "store2")
			},
		},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			db := testDB(t, func(db *Database) {
				_, err := db.CreateObjectStore("store1", ObjectStoreOptions{})
				assert.NoError(t, err)
				_, err = db.CreateObjectStore("store2", ObjectStoreOptions{})
				assert.NoError(t, err)
			})
			// set values in 2 stores
			txn, err := tc.makeTxn(db)
			assert.NoError(t, err)
			store1, err := txn.ObjectStore("store1")
			assert.NoError(t, err)
			_, err = store1.PutKey(js.ValueOf("key1"), js.ValueOf("value1"))
			assert.NoError(t, err)
			store2, err := txn.ObjectStore("store2")
			assert.NoError(t, err)
			_, err = store2.PutKey(js.ValueOf("key2"), js.ValueOf("value2"))
			assert.NoError(t, err)

			// verify 1 of the values is correct
			req, err := store1.GetAllKeys()
			assert.NoError(t, err)

			// wait for the whole txn to complete
			assert.NoError(t, txn.Await(context.Background()))
			result, err := req.Result()
			assert.NoError(t, err)
			assert.Equal(t, []js.Value{js.ValueOf("key1")}, result)
		})
	}
}

func TestDatabaseClose(t *testing.T) {
	t.Parallel()

	db := testDB(t, func(db *Database) {
		_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
		assert.NoError(t, err)
	})
	_, err := db.Transaction(TransactionReadOnly, "mystore")
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	_, err = db.Transaction(TransactionReadOnly, "mystore")
	assert.Error(t, err)
}
