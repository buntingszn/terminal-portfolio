//go:build js

package tea

import (
	"errors"
	"os"
)

// initInput is a no-op in WASM; input comes from the JavaScript bridge.
func (p *Program) initInput() error {
	return nil
}

// openInputTTY returns an error in WASM since there is no TTY.
func openInputTTY() (*os.File, error) {
	return nil, errors.New("TTY not available in WASM")
}

// suspendSupported is false in WASM.
const suspendSupported = false

// suspendProcess is a no-op in WASM.
func suspendProcess() {}
