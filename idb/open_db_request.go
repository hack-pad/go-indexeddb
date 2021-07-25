// +build js,wasm

package idb

import (
	"context"
	"fmt"
	"log"
	"syscall/js"
)

// OpenDBRequest provides access to the results of requests to open or delete databases (performed using Factory.open and Factory.DeleteDatabase).
type OpenDBRequest struct {
	*Request
}

// Upgrader is a function that can upgrade the given database from an old version to a new one.
type Upgrader func(db *Database, oldVersion, newVersion uint) error

func newOpenDBRequest(ctx context.Context, req *Request, upgrader Upgrader) *OpenDBRequest {
	ctx, cancel := context.WithCancel(ctx)
	req.ListenSuccess(ctx, func() {
		defer cancel()
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
	upgrade := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		var err error
		jsDatabase, err := req.Result()
		if err != nil {
			panic(err)
		}
		db := wrapDatabase(jsDatabase)
		oldVersion, newVersion := event.Get("oldVersion").Int(), event.Get("newVersion").Int()
		if oldVersion < 0 || newVersion < 0 {
			panic(fmt.Errorf("Unexpected negative oldVersion or newVersion: %d, %d", oldVersion, newVersion))
		}
		err = upgrader(db, uint(oldVersion), uint(newVersion))
		if err != nil {
			panic(err)
		}
		return nil
	})
	req.jsRequest.Call(addEventListener, "upgradeneeded", upgrade)
	go func() {
		<-ctx.Done()
		req.jsRequest.Call(removeEventListener, "upgradeneeded", upgrade)
		upgrade.Release()
	}()
	return &OpenDBRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (o *OpenDBRequest) Result() (*Database, error) {
	db, err := o.Request.Result()
	if err != nil {
		return nil, err
	}
	return wrapDatabase(db), nil
}

// Await waits for success or failure, then returns the results.
func (o *OpenDBRequest) Await(ctx context.Context) (*Database, error) {
	db, err := o.Request.Await(ctx)
	if err != nil {
		return nil, err
	}
	return wrapDatabase(db), nil
}
