// +build js,wasm

package idb

import (
	"syscall/js"
	"testing"

	"github.com/hack-pad/go-indexeddb/idb/internal/assert"
)

func TestGlobal(t *testing.T) {
	dbFactory, err := Global()
	assert.NoError(t, err)
	assert.Equal(t, &Factory{js.Global().Get("indexedDB")}, dbFactory)
}

func testFactory(tb testing.TB) *Factory {
	tb.Helper()
	dbFactory, err := Global()
	if !assert.NoError(tb, err) {
		tb.FailNow()
	}
	tb.Cleanup(func() {
		databaseNames := testGetDatabases(tb, dbFactory)
		var requests []*AckRequest
		for _, name := range databaseNames {
			req, err := dbFactory.DeleteDatabase(name)
			assert.NoError(tb, err)
			requests = append(requests, req)
		}
		for _, req := range requests {
			assert.NoError(tb, req.Await())
		}
	})
	return dbFactory
}

func testGetDatabases(tb testing.TB, dbFactory *Factory) []string {
	tb.Helper()
	done := make(chan struct{})
	var names []string
	var fn js.Func
	fn = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer fn.Release()
		arr := args[0]
		assert.NoError(tb, iterArray(arr, func(i int, value js.Value) (keepGoing bool) {
			names = append(names, value.Get("name").String())
			return true
		}))
		close(done)
		return nil
	})
	dbFactory.jsFactory.Call("databases").Call("then", fn)
	<-done
	return names
}

func TestFactoryOpen(t *testing.T) {
	t.Run("open new DB", func(t *testing.T) {
		dbFactory := testFactory(t)
		req, err := dbFactory.Open("mydb", 0, func(db *Database, oldVersion, newVersion uint) error {
			assert.Equal(t, uint(0), oldVersion)
			assert.Equal(t, uint(1), newVersion)
			return nil
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		db, err := req.Await()
		assert.NoError(t, err)
		assert.NotZero(t, db)
	})

	t.Run("open existing DB", func(t *testing.T) {
		dbFactory := testFactory(t)
		_, err := dbFactory.Open("mydb", 1, func(db *Database, oldVersion, newVersion uint) error {
			return nil
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		req, err := dbFactory.Open("mydb", 1, func(db *Database, oldVersion, newVersion uint) error {
			t.Error("Should not call upgrade")
			return nil
		})
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		db, err := req.Await()
		assert.NoError(t, err)
		assert.NotZero(t, db)
	})
}

func TestFactoryDeleteDatabase(t *testing.T) {
	t.Run("missing DB", func(t *testing.T) {
		dbFactory := testFactory(t)
		req, err := dbFactory.DeleteDatabase("does not exist")
		assert.NoError(t, err)
		err = req.Await()
		assert.NoError(t, err)
	})

	t.Run("delete DB", func(t *testing.T) {
		dbFactory := testFactory(t)
		var db *Database
		{
			req, err := dbFactory.Open("mydb", 0, func(db *Database, oldVersion, newVersion uint) error {
				_, err := db.CreateObjectStore("mystore", ObjectStoreOptions{})
				assert.NoError(t, err)
				return nil
			})
			assert.NoError(t, err)
			db, err = req.Await()
			assert.NoError(t, err)
			names, err := db.ObjectStoreNames()
			assert.NoError(t, err)
			assert.Equal(t, []string{"mystore"}, names)
			if t.Failed() {
				t.FailNow()
			}
		}

		req, err := dbFactory.DeleteDatabase("mydb")
		assert.NoError(t, err)
		err = req.Await()
		assert.NoError(t, err)

		// database should be closed and unusable now
		_, err = db.Transaction(TransactionReadOnly, "mystore")
		assert.Error(t, err)
	})
}

func TestFactoryCompareKeys(t *testing.T) {
	t.Run("normal keys", func(t *testing.T) {
		dbFactory := testFactory(t)
		compare, err := dbFactory.CompareKeys(js.ValueOf("a"), js.ValueOf("b"))
		assert.NoError(t, err)
		assert.Equal(t, -1, compare)
	})

	t.Run("bad keys", func(t *testing.T) {
		dbFactory := testFactory(t)
		_, err := dbFactory.CompareKeys(js.ValueOf(map[string]interface{}{"a": "a"}), js.ValueOf("b"))
		assert.Error(t, err)
	})
}
