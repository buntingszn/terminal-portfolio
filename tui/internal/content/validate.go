package content

import "fmt"

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
	if err := requireField("location", a.Location); err != nil {
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
