.PHONY: check validate wasm

check: validate
	@if [ -f web/package.json ]; then cd web && npm run check && npm run lint && npm run build; fi
	@if [ -f tui/Makefile ]; then cd tui && make check; fi

validate:
	@if [ -f scripts/validate-data.sh ]; then ./scripts/validate-data.sh; fi

wasm:
	cp data/content/*.json tui/internal/content/embed/
	cd tui && GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" \
		-o ../web/public/main.wasm ./cmd/wasm
	@GOROOT=$$(cd tui && go env GOROOT); \
	if [ -f "$$GOROOT/misc/wasm/wasm_exec.js" ]; then \
		cp "$$GOROOT/misc/wasm/wasm_exec.js" web/public/; \
	elif [ -f "$$GOROOT/lib/wasm/wasm_exec.js" ]; then \
		cp "$$GOROOT/lib/wasm/wasm_exec.js" web/public/; \
	else \
		echo "ERROR: wasm_exec.js not found in GOROOT ($$GOROOT)"; exit 1; \
	fi
