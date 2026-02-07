//go:build !js

package content

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadAll reads and validates all JSON data files from the given data directory.
// The dataDir should point to the root data/ directory containing a content/ subdirectory.
func LoadAll(dataDir string) (*Content, error) {
	contentDir := filepath.Join(dataDir, "content")

	info, err := os.Stat(contentDir)
	if err != nil {
		return nil, fmt.Errorf("content directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("content path is not a directory: %s", contentDir)
	}

	var c Content

	// Load meta.json
	if err := loadJSON(filepath.Join(contentDir, "meta.json"), &c.Meta); err != nil {
		return nil, fmt.Errorf("loading meta.json: %w", err)
	}
	if err := validateMeta(&c.Meta); err != nil {
		return nil, fmt.Errorf("meta.json: %w", err)
	}

	// Load about.json
	if err := loadJSON(filepath.Join(contentDir, "about.json"), &c.About); err != nil {
		return nil, fmt.Errorf("loading about.json: %w", err)
	}
	if err := validateAbout(&c.About); err != nil {
		return nil, fmt.Errorf("about.json: %w", err)
	}

	// Load work.json
	if err := loadJSON(filepath.Join(contentDir, "work.json"), &c.Work); err != nil {
		return nil, fmt.Errorf("loading work.json: %w", err)
	}
	if err := validateWork(&c.Work); err != nil {
		return nil, fmt.Errorf("work.json: %w", err)
	}

	// Load cv.json
	if err := loadJSON(filepath.Join(contentDir, "cv.json"), &c.CV); err != nil {
		return nil, fmt.Errorf("loading cv.json: %w", err)
	}
	if err := validateCV(&c.CV); err != nil {
		return nil, fmt.Errorf("cv.json: %w", err)
	}

	// Load links.json
	if err := loadJSON(filepath.Join(contentDir, "links.json"), &c.Links); err != nil {
		return nil, fmt.Errorf("loading links.json: %w", err)
	}
	if err := validateLinks(&c.Links); err != nil {
		return nil, fmt.Errorf("links.json: %w", err)
	}

	return &c, nil
}

// loadJSON reads a JSON file from disk and unmarshals it into v.
func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filepath.Base(path), err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
	}
	return nil
}

