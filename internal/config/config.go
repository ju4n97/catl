package config

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

// File models the root structure of a catl.toml file.
type File struct {
	Catalog      *Catalog              `toml:"catalog,omitzero"`
	Workspace    *Workspace            `toml:"workspace,omitzero"`
	Collections  map[string]Collection `toml:"collections,omitzero"`
	Dependencies map[string]Dependency `toml:"dependencies,omitzero"`
}

// Catalog names an individual catalog.
type Catalog struct {
	Name string `toml:"name"`
}

// Workspace groups several catalog members in one repository.
type Workspace struct {
	Members []string `toml:"members"`
}

// Collection configures the schema and storage directory of a collection.
type Collection struct {
	Schema string `toml:"schema,omitzero"`
	Dir    string `toml:"dir,omitzero"`
}

// Dependency declares a direct external catalog source.
// It is a pointer union keyed by variant name.
//
//	git     = { repo = "https://...", tag = "v1" }
//	archive = { url = "https://...", digest = "sha256:..." }
//	path    = { dir = "../core" }
//
// Exactly one field must be set. Use Source to get the validated variant.
type Dependency struct {
	Git     *GitSource     `toml:"git,omitzero"`
	Archive *ArchiveSource `toml:"archive,omitzero"`
	Path    *PathSource    `toml:"path,omitzero"`
}

// Source is a sealed sum type: GitSource, ArchiveSource or PathSource.
//
//sumtype:decl
type Source interface {
	validate() error
}

// GitSource is a remote Git repository.
type GitSource struct {
	Repo   string `toml:"repo"`
	Tag    string `toml:"tag,omitzero"`
	Branch string `toml:"branch,omitzero"`
	Rev    string `toml:"rev,omitzero"`
	Subdir string `toml:"subdir,omitzero"`
}

// validate implements [Source].
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

// ArchiveSource is a remote HTTPS tar archive.
type ArchiveSource struct {
	URL    string `toml:"url"`
	Digest string `toml:"digest,omitzero"`
}

// validate implements [Source].
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
	Dir string `toml:"dir"`
}

// validate implements [Source].
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

// Source returns the validated variant, or an error unless exactly one
// variant is set and valid.
func (d *Dependency) Source() (Source, error) {
	var (
		src   Source
		names []string
	)
	if d.Git != nil {
		src, names = *d.Git, append(names, "git")
	}
	if d.Archive != nil {
		src, names = *d.Archive, append(names, "archive")
	}
	if d.Path != nil {
		src, names = *d.Path, append(names, "path")
	}

	switch len(names) {
	case 0:
		return nil, errors.New("must specify one of 'git', 'archive', or 'path'")
	case 1:
		if err := src.validate(); err != nil {
			return nil, err
		}
		return src, nil
	default:
		return nil, fmt.Errorf("multiple sources specified (%s); choose exactly one", strings.Join(names, ", "))
	}
}

// NewDependency wraps a Source for programmatic construction (e.g. `catl add`).
func NewDependency(src Source) (Dependency, error) {
	switch s := src.(type) {
	case GitSource:
		return Dependency{Git: &s}, nil
	case ArchiveSource:
		return Dependency{Archive: &s}, nil
	case PathSource:
		return Dependency{Path: &s}, nil
	default:
		return Dependency{}, fmt.Errorf("unsupported source type %T", src)
	}
}

// Clone returns a deep copy of a dependency.
func (d Dependency) Clone() Dependency {
	return Dependency{
		Git:     clonePtr(d.Git),
		Archive: clonePtr(d.Archive),
		Path:    clonePtr(d.Path),
	}
}

// clonePtr returns a deep copy of a pointer.
func clonePtr[T any](p *T) *T {
	if p == nil {
		return nil
	}
	return new(*p)
}

// countSet returns the number of non-empty strings in the given slice.
func countSet(ss ...string) (n int) {
	for _, s := range ss {
		if s != "" {
			n++
		}
	}
	return n
}
