# Changelog

## Unreleased

### Added

- Add `GetAll`, `GetAllKey`, and `GetAllRange` methods on `ObjectStore` and `Index` to retrieve multiple values without cursor iteration.

### Fixed

- Avoid a nil-pointer panic in `ObjectStore.GetAllKeysRange` and `Index.GetAllKeysRange` when `query` is `nil`. Passing `query == nil` now applies no key filter (while still honoring `maxCount` when `> 0`), matching the IndexedDB specification.
