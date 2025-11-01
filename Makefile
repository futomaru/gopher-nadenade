WEB_DIR := web
WASM_OUT := $(WEB_DIR)/main.wasm
WASM_EXEC := $(WEB_DIR)/wasm_exec.js
GOCACHE := $(CURDIR)/.cache/go-build

.PHONY: wasm serve clean

wasm:
	@echo "==> Building WASM..."
	mkdir -p $(GOCACHE) $(WEB_DIR)
	GOCACHE=$(GOCACHE) GOOS=js GOARCH=wasm go build -o $(WASM_OUT) ./wasm
	cp $$(go env GOROOT)/lib/wasm/wasm_exec.js $(WASM_EXEC)
	@echo "✅ Build complete: $(WASM_OUT)"

serve:
	mkdir -p $(GOCACHE)
	GOCACHE=$(GOCACHE) go run ./cmd/server

clean:
	rm -f $(WASM_OUT) $(WASM_EXEC)
	rm -rf $(GOCACHE)
