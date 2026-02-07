.PHONY: check validate

check: validate
	@if [ -f web/package.json ]; then cd web && npm run check && npm run lint && npm run build; fi
	@if [ -f tui/Makefile ]; then cd tui && make check; fi

validate:
	@if [ -f scripts/validate-data.sh ]; then ./scripts/validate-data.sh; fi
