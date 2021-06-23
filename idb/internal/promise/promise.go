// +build js,wasm

package promise

// Promise is a JS-like promise interface. Enables alternative implementations, like Go channels.
type Promise interface {
	Then(fn func(value interface{}) interface{}) Promise
	Catch(fn func(value interface{}) interface{}) Promise
	Await() (interface{}, error)
}

// Resolver can be returned from a new Promise constructor, great for resolving successes and failures with values.
type Resolver func(interface{})
