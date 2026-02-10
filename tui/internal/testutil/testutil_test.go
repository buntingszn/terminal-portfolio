package testutil

import (
	"testing"
)

func TestFixtureContent_NonNil(t *testing.T) {
	c := FixtureContent()
	if c == nil {
		t.Fatal("FixtureContent() returned nil")
	}
}

func TestFixtureContent_MetaPopulated(t *testing.T) {
	c := FixtureContent()
	RequireNotEmpty(t, c.Meta.Name)
	RequireNotEmpty(t, c.Meta.Title)
	RequireNotEmpty(t, c.Meta.OneLiner)
	RequireNotEmpty(t, c.Meta.Version)
	RequireNotEmpty(t, c.Meta.SiteURL)
	RequireNotEmpty(t, c.Meta.SSHAddress)
	RequireNotEmpty(t, c.Meta.SourceRepo)
	RequireContains(t, c.Meta.Name, "Kyle McCormick")
}

func TestFixtureContent_AboutPopulated(t *testing.T) {
	c := FixtureContent()
	RequireNotEmpty(t, c.About.Bio)
	RequireNotEmpty(t, c.About.Email)
	RequireNotEmpty(t, c.About.Status)

	if len(c.About.Education) == 0 {
		t.Fatal("expected at least one education entry in About")
	}
}

func TestFixtureContent_WorkPopulated(t *testing.T) {
	c := FixtureContent()
	if len(c.Work.Projects) != 4 {
		t.Fatalf("expected 4 projects, got %d", len(c.Work.Projects))
	}

	// Verify first project has all fields
	p := c.Work.Projects[0]
	RequireNotEmpty(t, p.Title)
	RequireNotEmpty(t, p.Description)
	if len(p.Tags) == 0 {
		t.Fatal("expected at least one tag on first project")
	}

	// Verify there is at least one project with a URL or repo.
	hasURL := false
	for _, proj := range c.Work.Projects {
		if proj.URL != "" || proj.Repo != "" {
			hasURL = true
			break
		}
	}
	if !hasURL {
		t.Fatal("expected at least one project with a URL or repo")
	}
}

func TestFixtureContent_CVPopulated(t *testing.T) {
	c := FixtureContent()
	RequireNotEmpty(t, c.CV.Contact.Email)
	RequireNotEmpty(t, c.CV.Contact.Location)
	RequireNotEmpty(t, c.CV.Summary)

	if len(c.CV.Experience) != 2 {
		t.Fatalf("expected 2 experience entries, got %d", len(c.CV.Experience))
	}
	for i, exp := range c.CV.Experience {
		RequireNotEmpty(t, exp.Company)
		RequireNotEmpty(t, exp.Role)
		RequireNotEmpty(t, exp.Start)
		RequireNotEmpty(t, exp.End)
		if len(exp.Bullets) == 0 {
			t.Fatalf("expected bullets for experience entry %d", i)
		}
	}

	if len(c.CV.Skills) != 6 {
		t.Fatalf("expected 6 skill categories, got %d", len(c.CV.Skills))
	}
	for i, skill := range c.CV.Skills {
		RequireNotEmpty(t, skill.Category)
		if len(skill.Items) == 0 {
			t.Fatalf("expected items for skill category %d", i)
		}
	}

	if len(c.CV.Education) == 0 {
		t.Fatal("expected at least one education entry in CV")
	}
}

func TestFixtureContent_LinksPopulated(t *testing.T) {
	c := FixtureContent()
	if len(c.Links.Links) != 3 {
		t.Fatalf("expected 3 links, got %d", len(c.Links.Links))
	}
	for i, link := range c.Links.Links {
		RequireNotEmpty(t, link.Label)
		RequireNotEmpty(t, link.URL)
		RequireNotEmpty(t, link.Icon)
		if i == 0 {
			RequireContains(t, link.Label, "GitHub")
		}
	}
}

func TestFixtureTheme_ColorsSet(t *testing.T) {
	theme := FixtureTheme()
	if theme.Colors.Bg == "" {
		t.Fatal("expected Bg color to be set")
	}
	if theme.Colors.Fg == "" {
		t.Fatal("expected Fg color to be set")
	}
	if theme.Colors.Accent == "" {
		t.Fatal("expected Accent color to be set")
	}
	if theme.Colors.Muted == "" {
		t.Fatal("expected Muted color to be set")
	}
	if theme.Colors.Border == "" {
		t.Fatal("expected Border color to be set")
	}
}

func TestRequireContains_Pass(t *testing.T) {
	// Should not panic or fail for a valid substring match.
	RequireContains(t, "hello world", "world")
	RequireContains(t, "foobar", "foo")
	RequireContains(t, "exact", "exact")
}

func TestRequireContains_Fail(t *testing.T) {
	// Use a mock T to verify that RequireContains reports failure.
	mt := &testing.T{}
	RequireContains(mt, "hello", "xyz")
	if !mt.Failed() {
		t.Error("expected RequireContains to fail when substring is missing")
	}
}

func TestRequireNotEmpty_Pass(t *testing.T) {
	RequireNotEmpty(t, "non-empty")
	RequireNotEmpty(t, " ")
}

func TestRequireNotEmpty_Fail(t *testing.T) {
	mt := &testing.T{}
	RequireNotEmpty(mt, "")
	if !mt.Failed() {
		t.Error("expected RequireNotEmpty to fail on empty string")
	}
}
