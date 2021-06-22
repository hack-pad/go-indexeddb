BROWSERTEST_VERSION = v0.3.5
GO_BIN = $(shell printf '%s/bin' "$$(go env GOPATH)")

.PHONY: all
all: test

.PHONY: test-setup
test-setup:
	@if [ ! -f "${GO_BIN}/go_js_wasm_exec" ]; then \
		set -ex; \
		go install github.com/agnivade/wasmbrowsertest@${BROWSERTEST_VERSION}; \
		ln -s "${GO_BIN}/wasmbrowsertest" "${GO_BIN}/go_js_wasm_exec"; \
	fi

.PHONY: test
test: test-setup
	GOOS=js GOARCH=wasm go test -coverprofile=cover.out ./...
	go test -race ./...
