// +build js,wasm

package idb

import (
	"fmt"
	"math/rand"
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

	name := fmt.Sprintf("%s/%d", tb.Name(), rand.Int())
	req, err := dbFactory.Open(name, 0, func(db *Database, oldVersion, newVersion uint) error {
		initFunc(db)
		return nil
	})
	if !assert.NoError(tb, err) {
		tb.FailNow()
	}
	tb.Cleanup(func() {
		req, err := dbFactory.DeleteDatabase(name)
		assert.NoError(tb, err)
		assert.NoError(tb, req.Await())
	})
	db, err := req.Await()
	if !assert.NoError(tb, err) {
		tb.FailNow()
	}
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
		db := testDB(t, func(db *Database) {
			_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
			assert.NoError(t, err)
		})
		names, err := db.ObjectStoreNames()
		assert.NoError(t, err)
		assert.Equal(t, []string{"mystore"}, names)
	})

	t.Run("set keypath and auto-increment", func(t *testing.T) {
		const storeName = "mystore"
		db := testDB(t, func(db *Database) {
			_, err := db.CreateObjectStore(storeName, ObjectStoreOptions{
				KeyPath:       "primary",
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
