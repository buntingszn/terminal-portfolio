//go:build js

package config

// Load returns hardcoded defaults for WASM — no SSH, no rate limits,
// no idle timeout.
func Load() (*Config, error) {
	return &Config{}, nil
}
