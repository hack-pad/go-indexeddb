// +build js,wasm

package idb

import "syscall/js"

type Database struct {
	jsDB js.Value
}

func wrapDatabase(jsDB js.Value) (*Database, error) {
	return &Database{jsDB}, nil
}
