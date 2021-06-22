// +build js,wasm

package idb

import (
	"log"
	"syscall/js"

	"github.com/pkg/errors"
)

type OpenDBRequest struct {
	*Request
}

type Upgrader func(db *Database, oldVersion, newVersion uint) error

func wrapOpenDBRequest(req *Request, upgrader Upgrader) *OpenDBRequest {
	req.ListenSuccess(func() {
		jsDB, err := req.Result()
		if err != nil {
			panic(err)
		}
		jsDB.Call(addEventListener, "versionchange", js.FuncOf(func(js.Value, []js.Value) interface{} {
			log.Println("Version change detected, closing DB...")
			jsDB.Call("close")
			return nil
		}))
	})
	req.jsRequest.Call(addEventListener, "upgradeneeded", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		var err error
		jsDatabase, err := req.Result()
		if err != nil {
			panic(err)
		}
		db := wrapDatabase(jsDatabase)
		oldVersion, newVersion := event.Get("oldVersion").Int(), event.Get("newVersion").Int()
		if oldVersion < 0 || newVersion < 0 {
			panic(errors.Errorf("Unexpected negative oldVersion or newVersion: %d, %d", oldVersion, newVersion))
		}
		err = upgrader(db, uint(oldVersion), uint(newVersion))
		if err != nil {
			panic(err)
		}
		return nil
	}))
	return &OpenDBRequest{req}
}

func (o *OpenDBRequest) Result() (*Database, error) {
	db, err := o.Request.Result()
	if err != nil {
		return nil, err
	}
	return wrapDatabase(db), nil
}

func (o *OpenDBRequest) Await() (*Database, error) {
	db, err := o.Request.Await()
	if err != nil {
		return nil, err
	}
	return wrapDatabase(db), nil
}
