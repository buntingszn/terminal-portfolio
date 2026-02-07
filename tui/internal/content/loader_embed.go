//go:build js

package content

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed embed/*.json
var embeddedContent embed.FS

// LoadAll reads and validates all JSON data from embedded files.
// The dataDir argument is ignored in WASM — content is compiled in.
func LoadAll(_ string) (*Content, error) {
	var c Content

	if err := loadEmbeddedJSON("embed/meta.json", &c.Meta); err != nil {
		return nil, fmt.Errorf("loading meta.json: %w", err)
	}
	if err := validateMeta(&c.Meta); err != nil {
		return nil, fmt.Errorf("meta.json: %w", err)
	}

	if err := loadEmbeddedJSON("embed/about.json", &c.About); err != nil {
		return nil, fmt.Errorf("loading about.json: %w", err)
	}
	if err := validateAbout(&c.About); err != nil {
		return nil, fmt.Errorf("about.json: %w", err)
	}

	if err := loadEmbeddedJSON("embed/work.json", &c.Work); err != nil {
		return nil, fmt.Errorf("loading work.json: %w", err)
	}
	if err := validateWork(&c.Work); err != nil {
		return nil, fmt.Errorf("work.json: %w", err)
	}

	if err := loadEmbeddedJSON("embed/cv.json", &c.CV); err != nil {
		return nil, fmt.Errorf("loading cv.json: %w", err)
	}
	if err := validateCV(&c.CV); err != nil {
		return nil, fmt.Errorf("cv.json: %w", err)
	}

	if err := loadEmbeddedJSON("embed/links.json", &c.Links); err != nil {
		return nil, fmt.Errorf("loading links.json: %w", err)
	}
	if err := validateLinks(&c.Links); err != nil {
		return nil, fmt.Errorf("links.json: %w", err)
	}

	return &c, nil
}

func loadEmbeddedJSON(name string, v any) error {
	data, err := embeddedContent.ReadFile(name)
	if err != nil {
		return fmt.Errorf("reading %s: %w", name, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parsing %s: %w", name, err)
	}
	return nil
}
