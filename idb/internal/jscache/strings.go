// +build js,wasm

package jscache

import (
	"syscall/js"
)

var (
	jsReflectGet = js.Global().Get("Reflect").Get("get")
)

// Strings caches encoding strings as js.Value's.
// String encoding today is quite CPU intensive, so caching commonly used strings helps with performance.
type Strings struct {
	cacher
}

// Value retrieves the js.Value for the given string
func (c *Strings) Value(s string) js.Value {
	return c.value(s, identityStringGetter{s}.value)
}

// GetProperty retrieves the given object's property, using a cached string value if available. Saves on the performance cost of 2 round trips to JS.
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
