package config

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

// File models the root structure of a catl.toml configuration file.
type File struct {
	Catalog      *Catalog              `toml:"catalog,omitzero" json:"catalog,omitzero"`
	Workspace    *Workspace            `toml:"workspace,omitzero" json:"workspace,omitzero"`
	Collections  map[string]Collection `toml:"collections,omitzero" json:"collections,omitzero"`
	Dependencies map[string]Dependency `toml:"dependencies,omitzero" json:"dependencies,omitzero"`
}

// Catalog names an individual catalog.
type Catalog struct {
	Name string `toml:"name" json:"name"`
}

// Workspace groups several catalog members in one repository.
type Workspace struct {
	Members []string `toml:"members" json:"members"`
}

// Collection configures the schema and storage directory of a collection.
type Collection struct {
	Schema string `toml:"schema,omitzero" json:"schema,omitzero"`
	Dir    string `toml:"dir,omitzero" json:"dir,omitzero"`
}

// Dependency declares a direct external catalog source using the compact inline form:
//
//	shared-infra = { git = "https://...", tag = "v1" }
//	policies     = { url = "https://...", digest = "sha256:..." }
//	core         = { path = "../core" }
//
// Exactly one source type must be configured. Use Source to obtain the validated variant.
type Dependency struct {
	Git    string `toml:"git,omitzero" json:"git,omitzero"`
	Tag    string `toml:"tag,omitzero" json:"tag,omitzero"`
	Branch string `toml:"branch,omitzero" json:"branch,omitzero"`
	Rev    string `toml:"rev,omitzero" json:"rev,omitzero"`
	Subdir string `toml:"subdir,omitzero" json:"subdir,omitzero"`

	URL    string `toml:"url,omitzero" json:"url,omitzero"`
	Digest string `toml:"digest,omitzero" json:"digest,omitzero"`

	Path string `toml:"path,omitzero" json:"path,omitzero"`
}

// Source is a sealed sum type: GitSource, ArchiveSource, or PathSource.
//
//sumtype:decl
type Source interface {
	isSource()
	validate() error
}

// GitSource is a remote Git repository dependency.
type GitSource struct {
	Repo   string `json:"repo"`
	Tag    string `json:"tag,omitzero"`
	Branch string `json:"branch,omitzero"`
	Rev    string `json:"rev,omitzero"`
	Subdir string `json:"subdir,omitzero"`
}

func (GitSource) isSource() {}

// validate implements [Source] by verifying repository presence and target ref exclusivity.
func (s GitSource) validate() error {
	if s.Repo == "" {
		return errors.New("git: 'repo' is required")
	}
	if countSet(s.Tag, s.Branch, s.Rev) != 1 {
		return errors.New("git: exactly one of 'tag', 'branch', or 'rev' is required")
	}
	if s.Subdir != "" {
		if err := ValidateRelativePath(s.Subdir); err != nil {
			return fmt.Errorf("git: invalid subdir: %w", err)
		}
	}
	return nil
}

// ArchiveSource is a remote HTTPS tar archive dependency.
type ArchiveSource struct {
	URL    string `json:"url"`
	Digest string `json:"digest,omitzero"`
}

func (ArchiveSource) isSource() {}

// validate implements [Source] by ensuring the archive endpoint uses secure HTTPS.
func (s ArchiveSource) validate() error {
	if s.URL == "" {
		return errors.New("archive: 'url' is required")
	}
	u, err := url.Parse(s.URL)
	switch {
	case err != nil:
		return fmt.Errorf("archive: invalid url: %w", err)
	case u.Scheme != "https" || u.Host == "":
		return fmt.Errorf("archive: url %q must be an absolute https URL", s.URL)
	}
	return nil
}

// PathSource is a catalog on the local filesystem.
type PathSource struct {
	Dir string `json:"dir"`
}

func (PathSource) isSource() {}

// validate implements [Source] by rejecting absolute paths and Windows drive prefixes.
func (p PathSource) validate() error {
	switch {
	case p.Dir == "":
		return errors.New("path: 'dir' is required")
	case strings.ContainsAny(p.Dir, "\\:\x00"):
		return fmt.Errorf("path: dir %q must use forward slashes and contain no ':' or NUL", p.Dir)
	case path.IsAbs(p.Dir):
		return fmt.Errorf("path: dir %q must be relative", p.Dir)
	}
	return nil
}

// Source resolves the flat Dependency fields into a validated concrete Source variant.
func (d Dependency) Source() (Source, error) {
	var (
		src   Source
		names []string
	)
	if d.Git != "" {
		src = GitSource{
			Repo:   d.Git,
			Tag:    d.Tag,
			Branch: d.Branch,
			Rev:    d.Rev,
			Subdir: d.Subdir,
		}
		names = append(names, "git")
	}
	if d.URL != "" {
		src = ArchiveSource{
			URL:    d.URL,
			Digest: d.Digest,
		}
		names = append(names, "url")
	}
	if d.Path != "" {
		src = PathSource{
			Dir: d.Path,
		}
		names = append(names, "path")
	}

	switch len(names) {
	case 0:
		return nil, errors.New("must specify one of 'git', 'url', or 'path'")
	case 1:
		if err := src.validate(); err != nil {
			return nil, err
		}
		return src, nil
	default:
		return nil, fmt.Errorf("multiple sources specified (%s); choose exactly one", strings.Join(names, ", "))
	}
}

// NewDependency converts a concrete Source variant into its flat Dependency representation.
func NewDependency(src Source) (Dependency, error) {
	switch s := src.(type) {
	case GitSource:
		return Dependency{
			Git:    s.Repo,
			Tag:    s.Tag,
			Branch: s.Branch,
			Rev:    s.Rev,
			Subdir: s.Subdir,
		}, nil
	case ArchiveSource:
		return Dependency{
			URL:    s.URL,
			Digest: s.Digest,
		}, nil
	case PathSource:
		return Dependency{
			Path: s.Dir,
		}, nil
	default:
		return Dependency{}, fmt.Errorf("unsupported source type %T", src)
	}
}

// countSet counts how many strings in the slice are non-empty.
func countSet(ss ...string) (n int) {
	for _, s := range ss {
		if s != "" {
			n++
		}
	}
	return n
}
