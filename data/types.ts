/**
 * TypeScript type definitions for the terminal-portfolio shared data layer.
 *
 * These interfaces mirror the JSON schemas in data/schemas/ and are consumed
 * by both the Astro web front-end and any tooling that operates on the
 * shared JSON content files in data/content/.
 */

// ---------------------------------------------------------------------------
// Meta
// ---------------------------------------------------------------------------

/** Global site metadata including identity, version, and connection details. */
export interface Meta {
  /** Semantic version of the data schema (e.g. "1.0.0"). */
  version: string;
  /** Full display name of the portfolio owner. */
  name: string;
  /** Professional title or role (e.g. "Software Engineer"). */
  title: string;
  /** Short tagline displayed on the home screen. */
  oneLiner: string;
  /** Public URL of the web portfolio. */
  siteUrl: string;
  /** Public URL of the browser-based TUI portfolio (e.g. "https://ssh.kpm.fyi"). */
  sshAddress: string;
  /** URL of the source code repository. */
  sourceRepo: string;
}

// ---------------------------------------------------------------------------
// Education (shared between About and Cv)
// ---------------------------------------------------------------------------

/** A single education entry, shared by the About and CV data models. */
export interface Education {
  /** Name of the school or university. */
  institution: string;
  /** Degree or certification earned. */
  degree: string;
  /** Graduation year. */
  year: string;
}

// ---------------------------------------------------------------------------
// About
// ---------------------------------------------------------------------------

/** Biographical information, availability status, contact, education, and interests. */
export interface About {
  /** Short biography paragraph. */
  bio: string;
  /** Current availability or professional status. */
  status: string;
  /** Contact email address. */
  email: string;
  /** SSH hostname for the CLI portfolio. */
  cli: string;
  /** List of educational credentials. */
  education?: Education[];
  /** List of professional or personal interests. */
  interests?: string[];
}

// ---------------------------------------------------------------------------
// Work / Projects
// ---------------------------------------------------------------------------

/** A single project entry in the portfolio. */
export interface WorkProject {
  /** Display name of the project. */
  title: string;
  /** Short description of the project. */
  description: string;
  /** Technology tags associated with the project. */
  tags: string[];
  /** Live URL of the project. Empty string if not publicly hosted. */
  url: string;
  /** Source repository URL. Empty string if not open-source. */
  repo: string;
  /** Whether the project is highlighted on the home screen. */
  featured: boolean;
}

/** Portfolio of projects with descriptions, tags, links, and featured status. */
export interface Work {
  /** List of portfolio projects. */
  projects: WorkProject[];
}

// ---------------------------------------------------------------------------
// CV / Resume
// ---------------------------------------------------------------------------

/** Primary contact information for the CV. */
export interface CvContact {
  /** Contact email address. */
  email: string;
  /** Geographic location or region. */
  location: string;
  /** Personal website URL. */
  website: string;
}

/** A single work experience entry. */
export interface CvExperience {
  /** Employer or organization name. */
  company: string;
  /** Job title or role. */
  role: string;
  /** Start date (year or year-month). */
  start: string;
  /** End date (year, year-month, or "Present"). */
  end: string;
  /** Accomplishment bullet points. */
  bullets: string[];
}

/** A skill category grouping related skills. */
export interface CvSkill {
  /** Name of the skill category (e.g. "Languages", "Backend"). */
  category: string;
  /** Individual skills within this category. */
  items: string[];
}

/** Structured resume data including contact info, experience, skills, and education. */
export interface Cv {
  /** Primary contact information. */
  contact: CvContact;
  /** Professional summary paragraph for the CV header. */
  summary: string;
  /** Work experience entries in reverse chronological order. */
  experience: CvExperience[];
  /** Skill categories with lists of individual skills. */
  skills: CvSkill[];
  /** Educational credentials. */
  education: Education[];
}

// ---------------------------------------------------------------------------
// Links
// ---------------------------------------------------------------------------

/** A single external link entry. */
export interface Link {
  /** Display label for the link. */
  label: string;
  /** Target URL (supports https, mailto, ssh, and other URI schemes). */
  url: string;
  /** Optional display text shown instead of the raw URL. */
  text?: string;
  /** Icon identifier used by both web and TUI renderers (e.g. "github", "mail", "terminal"). */
  icon: string;
}

/** External links for social profiles, contact, and other resources. */
export interface Links {
  /** List of external links. */
  links: Link[];
}
