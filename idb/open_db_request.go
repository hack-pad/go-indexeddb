//go:build js && wasm
// +build js,wasm

package idb

import (
	"context"
	"fmt"
	"log"

	"github.com/hack-pad/safejs"
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
		jsDB, err := req.safeResult()
		if err != nil {
			panic(err)
		}
		_, err = jsDB.Call(addEventListener, "versionchange", safejs.Must(safejs.FuncOf(func(safejs.Value, []safejs.Value) interface{} {
			log.Println("Version change detected, closing DB...")
			_, closeErr := jsDB.Call("close")
			if closeErr != nil {
				panic(closeErr)
			}
			return nil
		})))
		if err != nil {
			panic(err)
		}
	})
	upgrade, err := safejs.FuncOf(func(this safejs.Value, args []safejs.Value) interface{} {
		event := args[0]
		var err error
		jsDatabase, err := req.safeResult()
		if err != nil {
			panic(err)
		}
		db := wrapDatabase(jsDatabase)
		oldVersionValue, err := event.Get("oldVersion")
		if err != nil {
			panic(err)
		}
		oldVersion, err := oldVersionValue.Int()
		if err != nil {
			panic(err)
		}
		newVersionValue, err := event.Get("newVersion")
		if err != nil {
			panic(err)
		}
		newVersion, err := newVersionValue.Int()
		if err != nil {
			panic(err)
		}
		if oldVersion < 0 || newVersion < 0 {
			panic(fmt.Errorf("Unexpected negative oldVersion or newVersion: %d, %d", oldVersion, newVersion))
		}
		err = upgrader(db, uint(oldVersion), uint(newVersion))
		if err != nil {
			panic(err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	_, err = req.jsRequest.Call(addEventListener, "upgradeneeded", upgrade)
	if err != nil {
		panic(err)
	}
	go func() {
		<-ctx.Done()
		_, err := req.jsRequest.Call(removeEventListener, "upgradeneeded", upgrade)
		if err != nil {
			panic(err)
		}
		upgrade.Release()
	}()
	return &OpenDBRequest{req}
}

// Result returns the result of the request. If the request failed and the result is not available, an error is returned.
func (o *OpenDBRequest) Result() (*Database, error) {
	db, err := o.Request.safeResult()
	if err != nil {
		return nil, err
	}
	return wrapDatabase(db), nil
}

// Await waits for success or failure, then returns the results.
func (o *OpenDBRequest) Await(ctx context.Context) (*Database, error) {
	db, err := o.Request.safeAwait(ctx)
	if err != nil {
		return nil, err
	}
	return wrapDatabase(db), nil
}
