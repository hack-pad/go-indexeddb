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

func stringsFromArray(arr js.Value) (_ []string, err error) {
	defer exception.Catch(&err)
	length := arr.Length()
	strs := make([]string, length)
	for i := 0; i < length; i++ {
		strs[i] = arr.Index(i).String()
	}
	return strs, nil
}
