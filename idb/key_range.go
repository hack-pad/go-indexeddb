// +build js,wasm

package idb

import (
	"syscall/js"

	"github.com/hack-pad/go-indexeddb/idb/internal/exception"
)

var (
	jsIDBKeyRange = js.Global().Get("IDBKeyRange")
)

type KeyRange struct {
	jsKeyRange js.Value
}

func wrapKeyRange(jsKeyRange js.Value) *KeyRange {
	return &KeyRange{jsKeyRange}
}

func NewKeyRangeBound(lower, upper js.Value, lowerOpen, upperOpen bool) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("bound", lower, upper, lowerOpen, upperOpen)), nil
}

func NewKeyRangeLowerBound(lower js.Value, open bool) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("lowerBound", lower, open)), nil
}

func NewKeyRangeUpperBound(upper js.Value, open bool) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("upperBound", upper, open)), nil
}

func NewKeyRangeOnly(only js.Value) (_ *KeyRange, err error) {
	defer exception.Catch(&err)
	return wrapKeyRange(jsIDBKeyRange.Call("only", only)), nil
}

func (k *KeyRange) Lower() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return k.jsKeyRange.Get("lower"), nil
}

func (k *KeyRange) Upper() (_ js.Value, err error) {
	defer exception.Catch(&err)
	return k.jsKeyRange.Get("upper"), nil
}

func (k *KeyRange) LowerOpen() (_ bool, err error) {
	defer exception.Catch(&err)
	return k.jsKeyRange.Get("lowerOpen").Bool(), nil
}

func (k *KeyRange) UpperOpen() (_ bool, err error) {
	defer exception.Catch(&err)
	return k.jsKeyRange.Get("upperOpen").Bool(), nil
}

func (k *KeyRange) Includes(key js.Value) (_ bool, err error) {
	defer exception.Catch(&err)
	return k.jsKeyRange.Call("includes", key).Bool(), nil
}

func (k *KeyRange) JSValue() js.Value {
	return k.jsKeyRange
}
