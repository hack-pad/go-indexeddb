//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"strings"
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
	"github.com/hack-pad/safejs"
)

const testDBPrefix = "go-indexeddb-test-"

func TestGlobal(t *testing.T) {
	t.Parallel()
	var dbFactory *Factory
	assert.NotPanics(t, func() {
		dbFactory = Global()
	})
	assert.Equal(t, &Factory{safejs.Must(safejs.Global().Get("indexedDB"))}, dbFactory)
}

func testFactory(tb testing.TB) *Factory {
	tb.Helper()
	dbFactory := Global()
	tb.Cleanup(func() {
		databaseNames := testGetDatabases(tb, dbFactory)
		var requests []*AckRequest
		for _, name := range databaseNames {
			if strings.HasPrefix(name, testDBPrefix) {
				req, err := dbFactory.DeleteDatabase(name)
				assert.NoError(tb, err)
				requests = append(requests, req)
			}
		}
		for _, req := range requests {
			assert.NoError(tb, req.Await(context.Background()))
		}
	})
	return dbFactory
}

func testGetDatabases(tb testing.TB, dbFactory *Factory) []string {
	tb.Helper()
	done := make(chan struct{})
	var names []string
	var fn safejs.Func
	fn, err := safejs.FuncOf(func(_ safejs.Value, args []safejs.Value) interface{} {
		defer fn.Release()
		arr := args[0]
		assert.NoError(tb, iterArray(arr, func(_ int, value safejs.Value) (keepGoing bool, visitErr error) {
			name := safejs.Must(safejs.Must(value.Get("name")).String())
			names = append(names, name)
			return true, nil
		}))
		close(done)
		return nil
	})
	if err != nil {
		assert.NoError(tb, err)
	}
	databasesPromise, err := dbFactory.jsFactory.Call("databases")
	if err != nil {
		assert.NoError(tb, err)
	}
	_, err = databasesPromise.Call("then", fn)
	if err != nil {
		assert.NoError(tb, err)
	}
	<-done
	return names
}

func TestFactoryOpenNewDB(t *testing.T) { // nolint:paralleltest // Deletes all databases, should not run in parallel.
	dbFactory := testFactory(t)
	req, err := dbFactory.Open(context.Background(), testDBPrefix+"mydb", 0, func(db *Database, oldVersion, newVersion uint) error {
		assert.Equal(t, uint(0), oldVersion)
		assert.Equal(t, uint(1), newVersion)
		return nil
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	db, err := req.Await(context.Background())
	assert.NoError(t, err)
	assert.NotZero(t, db)
	assert.NoError(t, db.Close())
}

func TestFactoryOpenExistingDB(t *testing.T) { // nolint:paralleltest // Deletes all databases, should not run in parallel.
	dbFactory := testFactory(t)
	_, err := dbFactory.Open(context.Background(), testDBPrefix+"mydb", 1, func(db *Database, oldVersion, newVersion uint) error {
		return nil
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	req, err := dbFactory.Open(context.Background(), testDBPrefix+"mydb", 1, func(db *Database, oldVersion, newVersion uint) error {
		t.Error("Should not call upgrade")
		return nil
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	db, err := req.Await(context.Background())
	assert.NoError(t, err)
	assert.NotZero(t, db)
	assert.NoError(t, db.Close())
}

func TestFactoryDeleteMissingDatabase(t *testing.T) { // nolint:paralleltest // Deletes all databases, should not run in parallel.
	dbFactory := testFactory(t)
	req, err := dbFactory.DeleteDatabase("does not exist")
	assert.NoError(t, err)
	err = req.Await(context.Background())
	assert.NoError(t, err)
}

func TestFactoryDeleteDatabase(t *testing.T) { // nolint:paralleltest // Deletes all databases, should not run in parallel.
	dbFactory := testFactory(t)
	var db *Database
	{
		req, err := dbFactory.Open(context.Background(), testDBPrefix+"mydb", 0, func(db *Database, oldVersion, newVersion uint) error {
			_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
			assert.NoError(t, err)
			return nil
		})
		assert.NoError(t, err)
		db, err = req.Await(context.Background())
		assert.NoError(t, err)
		names, err := db.ObjectStoreNames()
		assert.NoError(t, err)
		assert.Equal(t, []string{"mystore"}, names)
		if t.Failed() {
			t.FailNow()
		}
	}

	req, err := dbFactory.DeleteDatabase(testDBPrefix + "mydb")
	assert.NoError(t, err)
	err = req.Await(context.Background())
	assert.NoError(t, err)

	// database should be closed and unusable now
	_, err = db.Transaction(TransactionReadOnly, "mystore")
	assert.Error(t, err)
	assert.NoError(t, db.Close())
}

func TestFactoryCompareKeys(t *testing.T) {
	t.Parallel()

	t.Run("normal keys", func(t *testing.T) {
		t.Parallel()
		dbFactory := testFactory(t)
		compare, err := dbFactory.CompareKeys(js.ValueOf("a"), js.ValueOf("b"))
		assert.NoError(t, err)
		assert.Equal(t, -1, compare)
	})

	t.Run("bad keys", func(t *testing.T) {
		t.Parallel()
		dbFactory := testFactory(t)
		_, err := dbFactory.CompareKeys(js.ValueOf(map[string]interface{}{"a": "a"}), js.ValueOf("b"))
		assert.Error(t, err)
	})
}
