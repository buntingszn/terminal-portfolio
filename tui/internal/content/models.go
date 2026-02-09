package content

// Meta holds site metadata from meta.json.
type Meta struct {
	Version    string `json:"version"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	OneLiner   string `json:"oneLiner"`
	SiteURL    string `json:"siteUrl"`
	SSHAddress string `json:"sshAddress"`
	SourceRepo string `json:"sourceRepo"`
}

// Education represents an education entry shared by About and CV.
type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Year        string `json:"year"`
}

// About holds bio and personal info from about.json.
type About struct {
	Bio       string      `json:"bio"`
	Location  string      `json:"location"`
	Status    string      `json:"status"`
	Education []Education `json:"education"`
	Interests []string    `json:"interests,omitempty"`
}

// WorkProject represents a single project entry.
type WorkProject struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	URL         string   `json:"url"`
	Repo        string   `json:"repo"`
	Featured    bool     `json:"featured"`
}

// Work holds the projects list from work.json.
type Work struct {
	Projects []WorkProject `json:"projects"`
}

// CVContact holds contact information.
type CVContact struct {
	Email    string `json:"email"`
	Location string `json:"location"`
	Website  string `json:"website"`
}

// CVExperience represents a work experience entry.
type CVExperience struct {
	Company string   `json:"company"`
	Role    string   `json:"role"`
	Start   string   `json:"start"`
	End     string   `json:"end"`
	Bullets []string `json:"bullets"`
}

// CVSkill represents a skill category with its items.
type CVSkill struct {
	Category string   `json:"category"`
	Items    []string `json:"items"`
}

// CV holds the full CV data from cv.json.
type CV struct {
	Contact    CVContact      `json:"contact"`
	Summary    string         `json:"summary"`
	Experience []CVExperience `json:"experience"`
	Skills     []CVSkill      `json:"skills"`
	Education  []Education    `json:"education"`
}

// Link represents an external link entry.
type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Icon  string `json:"icon"`
}

// Links holds the links list from links.json.
type Links struct {
	Links []Link `json:"links"`
}

// Content holds all loaded site data.
type Content struct {
	Meta  Meta
	About About
	Work  Work
	CV    CV
	Links Links
}
