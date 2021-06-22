// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

func sliceFromStrings(strs []string) []interface{} {
	values := make([]interface{}, 0, len(strs))
	for _, s := range strs {
		values = append(values, s)
	}
	return values
}

func stringsFromArray(arr js.Value) (strs []string, err error) {
	defer exception.Catch(&err)
	err = iterArray(arr, func(i int, value js.Value) bool {
		strs = append(strs, value.String())
		return true
	})
	return
}

func iterArray(arr js.Value, visit func(i int, value js.Value) (keepGoing bool)) (err error) {
	defer exception.Catch(&err)
	length := arr.Length()
	for i := 0; i < length; i++ {
		visit(i, arr.Index(i))
	}
	return nil
}
