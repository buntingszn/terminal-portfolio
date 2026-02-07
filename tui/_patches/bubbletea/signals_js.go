//go:build js

package tea

// listenForResize is a no-op in WASM; resize events come from xterm.js
// via Program.Send(WindowSizeMsg{...}).
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
