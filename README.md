# go-indexeddb    [![Go Reference](https://pkg.go.dev/badge/github.com/hack-pad/go-indexeddb/idb.svg)][reference] [![CI](https://github.com/hack-pad/go-indexeddb/actions/workflows/ci.yml/badge.svg)](https://github.com/hack-pad/go-indexeddb/actions/workflows/ci.yml)

An IndexedDB driver with bindings for Go code compiled to WebAssembly.

Package `idb` is a low-level Go driver that provides type-safe bindings to IndexedDB in Wasm programs.
The primary focus is to align with the IndexedDB spec, followed by ease of use.

To get started, get the global indexedDB instance with idb.Global(). See the [reference][] for examples and full documentation.

```bash
go get github.com/hack-pad/go-indexeddb@latest
```
```go
import "github.com/hack-pad/go-indexeddb/idb"
```

[reference]: https://pkg.go.dev/github.com/hack-pad/go-indexeddb/idb
