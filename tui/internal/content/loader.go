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

// requireField returns an error if value is empty.
func requireField(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

func validateMeta(m *Meta) error {
	if err := requireField("name", m.Name); err != nil {
		return err
	}
	if err := requireField("title", m.Title); err != nil {
		return err
	}
	if err := requireField("version", m.Version); err != nil {
		return err
	}
	return nil
}

func validateAbout(a *About) error {
	if err := requireField("bio", a.Bio); err != nil {
		return err
	}
	if err := requireField("email", a.Email); err != nil {
		return err
	}
	return nil
}

func validateWork(w *Work) error {
	if len(w.Projects) == 0 {
		return fmt.Errorf("projects list must not be empty")
	}
	for i, p := range w.Projects {
		if err := requireField("title", p.Title); err != nil {
			return fmt.Errorf("project[%d]: %w", i, err)
		}
		if err := requireField("description", p.Description); err != nil {
			return fmt.Errorf("project[%d]: %w", i, err)
		}
	}
	return nil
}

func validateCV(cv *CV) error {
	if err := requireField("summary", cv.Summary); err != nil {
		return err
	}
	if err := requireField("contact.email", cv.Contact.Email); err != nil {
		return err
	}
	if len(cv.Experience) == 0 {
		return fmt.Errorf("experience list must not be empty")
	}
	for i, e := range cv.Experience {
		if err := requireField("company", e.Company); err != nil {
			return fmt.Errorf("experience[%d]: %w", i, err)
		}
		if err := requireField("role", e.Role); err != nil {
			return fmt.Errorf("experience[%d]: %w", i, err)
		}
	}
	if len(cv.Skills) == 0 {
		return fmt.Errorf("skills list must not be empty")
	}
	return nil
}

func validateLinks(l *Links) error {
	if len(l.Links) == 0 {
		return fmt.Errorf("links list must not be empty")
	}
	for i, link := range l.Links {
		if err := requireField("label", link.Label); err != nil {
			return fmt.Errorf("link[%d]: %w", i, err)
		}
		if err := requireField("url", link.URL); err != nil {
			return fmt.Errorf("link[%d]: %w", i, err)
		}
	}
	return nil
}
