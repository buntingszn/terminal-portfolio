package content

import (
	"os"
	"path/filepath"
	"testing"
)

// dataDir returns the path to the shared data/ directory relative to this test file.
func dataDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("..", "..", "..", "data")
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("resolving data dir: %v", err)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("data directory not found at %s: %v", abs, err)
	}
	return abs
}

func TestLoadAll(t *testing.T) {
	c, err := LoadAll(dataDir(t))
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Meta checks
	if c.Meta.Name == "" {
		t.Error("Meta.Name is empty")
	}
	if c.Meta.Title == "" {
		t.Error("Meta.Title is empty")
	}
	if c.Meta.Version == "" {
		t.Error("Meta.Version is empty")
	}
	if c.Meta.SiteURL == "" {
		t.Error("Meta.SiteURL is empty")
	}

	// About checks
	if c.About.Bio == "" {
		t.Error("About.Bio is empty")
	}
	if c.About.Email == "" {
		t.Error("About.Email is empty")
	}

	// Work checks
	if len(c.Work.Projects) == 0 {
		t.Fatal("Work.Projects is empty")
	}
	for i, p := range c.Work.Projects {
		if p.Title == "" {
			t.Errorf("Work.Projects[%d].Title is empty", i)
		}
		if p.Description == "" {
			t.Errorf("Work.Projects[%d].Description is empty", i)
		}
	}

	// CV checks
	if c.CV.Summary == "" {
		t.Error("CV.Summary is empty")
	}
	if c.CV.Contact.Email == "" {
		t.Error("CV.Contact.Email is empty")
	}
	if len(c.CV.Experience) == 0 {
		t.Fatal("CV.Experience is empty")
	}
	for i, e := range c.CV.Experience {
		if e.Company == "" {
			t.Errorf("CV.Experience[%d].Company is empty", i)
		}
		if e.Role == "" {
			t.Errorf("CV.Experience[%d].Role is empty", i)
		}
		if len(e.Bullets) == 0 {
			t.Errorf("CV.Experience[%d].Bullets is empty", i)
		}
	}
	if len(c.CV.Skills) == 0 {
		t.Error("CV.Skills is empty")
	}

	// Links checks
	if len(c.Links.Links) == 0 {
		t.Fatal("Links.Links is empty")
	}
	for i, l := range c.Links.Links {
		if l.Label == "" {
			t.Errorf("Links.Links[%d].Label is empty", i)
		}
		if l.URL == "" {
			t.Errorf("Links.Links[%d].URL is empty", i)
		}
	}
}

func TestLoadAllInvalidDir(t *testing.T) {
	_, err := LoadAll("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestLoadAllMissingFile(t *testing.T) {
	// Create a temporary directory with a content/ subdirectory but no files.
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.Mkdir(contentDir, 0o755); err != nil {
		t.Fatalf("creating content dir: %v", err)
	}

	_, err := LoadAll(tmpDir)
	if err == nil {
		t.Fatal("expected error for missing JSON files")
	}
}

func TestLoadAllInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.Mkdir(contentDir, 0o755); err != nil {
		t.Fatalf("creating content dir: %v", err)
	}

	// Write invalid JSON to meta.json.
	if err := os.WriteFile(filepath.Join(contentDir, "meta.json"), []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("writing meta.json: %v", err)
	}

	_, err := LoadAll(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadAllValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.Mkdir(contentDir, 0o755); err != nil {
		t.Fatalf("creating content dir: %v", err)
	}

	// Write meta.json with missing required name field.
	metaJSON := `{"version":"1.0.0","name":"","title":"Engineer","oneLiner":"test"}`
	if err := os.WriteFile(filepath.Join(contentDir, "meta.json"), []byte(metaJSON), 0o644); err != nil {
		t.Fatalf("writing meta.json: %v", err)
	}

	_, err := LoadAll(tmpDir)
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}
}

func TestLoadAllWorkValidation(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.Mkdir(contentDir, 0o755); err != nil {
		t.Fatalf("creating content dir: %v", err)
	}

	// Valid meta.json
	writeFile(t, contentDir, "meta.json", `{"version":"1.0.0","name":"Test","title":"Dev"}`)
	// Valid about.json
	writeFile(t, contentDir, "about.json", `{"bio":"A bio","email":"test@example.com","status":"Available"}`)
	// work.json with empty projects
	writeFile(t, contentDir, "work.json", `{"projects":[]}`)

	_, err := LoadAll(tmpDir)
	if err == nil {
		t.Fatal("expected validation error for empty projects list")
	}
}

func TestLoadAllLinksValidation(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content")
	if err := os.Mkdir(contentDir, 0o755); err != nil {
		t.Fatalf("creating content dir: %v", err)
	}

	writeFile(t, contentDir, "meta.json", `{"version":"1.0.0","name":"Test","title":"Dev"}`)
	writeFile(t, contentDir, "about.json", `{"bio":"A bio","email":"test@example.com","status":"Available"}`)
	writeFile(t, contentDir, "work.json", `{"projects":[{"title":"P","description":"D","tags":[],"url":"","repo":"","featured":false}]}`)
	writeFile(t, contentDir, "cv.json", `{"contact":{"email":"a@b.c","location":"X","website":"https://x"},"summary":"S","experience":[{"company":"C","role":"R","start":"2020","end":"2024","bullets":["b"]}],"skills":[{"category":"C","items":["i"]}],"education":[]}`)
	// links.json with missing label
	writeFile(t, contentDir, "links.json", `{"links":[{"label":"","url":"https://example.com","icon":"x"}]}`)

	_, err := LoadAll(tmpDir)
	if err == nil {
		t.Fatal("expected validation error for empty link label")
	}
}

func TestLoadAllContentFields(t *testing.T) {
	c, err := LoadAll(dataDir(t))
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// Verify specific values from the real data files.
	if c.Meta.Name != "Kyle McCormick" {
		t.Errorf("Meta.Name = %q, want %q", c.Meta.Name, "Kyle McCormick")
	}
	if c.About.Email != "hi@kpm.fyi" {
		t.Errorf("About.Email = %q, want %q", c.About.Email, "hi@kpm.fyi")
	}
	if c.CV.Contact.Email != "hi@kpm.fyi" {
		t.Errorf("CV.Contact.Email = %q, want %q", c.CV.Contact.Email, "hi@kpm.fyi")
	}

	// Verify projects loaded.
	if len(c.Work.Projects) == 0 {
		t.Error("expected at least one project")
	}
}

// writeFile is a test helper that writes content to a file in the given directory.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}
