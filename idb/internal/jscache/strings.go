// +build js,wasm

package jscache

import (
	"syscall/js"
)

var (
	jsReflectGet = js.Global().Get("Reflect").Get("get")
)

type Strings struct {
	cacher
}

func (c *Strings) Value(s string) js.Value {
	return c.value(s, identityStringGetter{s}.value)
}

func (c *Strings) GetProperty(obj js.Value, key string) js.Value {
	jsKey := c.Value(key)
	return jsReflectGet.Invoke(obj, jsKey)
}

type identityStringGetter struct {
	s string
}

func (i identityStringGetter) value() interface{} {
	return i.s
}
