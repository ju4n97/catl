package conf

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// File models the root structure of a catl.toml configuration file.
type File struct {
	Catalog      Catalog               `toml:"catalog,omitzero"      json:"catalog,omitzero"`
	Collections  map[string]Collection `toml:"collections,omitzero"  json:"collections,omitzero"`
	Dependencies map[string]Dependency `toml:"dependencies,omitzero" json:"dependencies,omitzero"`
}

// ValidateCollections checks relative path invariants and storage directory
// uniqueness across declared collections.
func (f *File) ValidateCollections() error {
	seenDirs := make(map[string]string, len(f.Collections))

	for name, coll := range f.Collections {
		if coll.Schema != "" {
			if err := ValidateRelativePath(coll.Schema); err != nil {
				return fmt.Errorf("collection %q schema: %w", name, err)
			}
		}

		dir := coll.Dir
		if dir == "" {
			dir = name
		} else {
			if err := ValidateRelativePath(dir); err != nil {
				return fmt.Errorf("collection %q dir: %w", name, err)
			}
		}

		if existing, exists := seenDirs[dir]; exists {
			return fmt.Errorf(
				"collections %q and %q share the same directory %q",
				existing,
				name,
				dir,
			)
		}
		seenDirs[dir] = name
	}

	return nil
}

// Catalog defines the primary identity and namespace of a catalog.
type Catalog struct {
	Name string `toml:"name" json:"name"`
}

// Collection configures the schema and storage directory of a collection.
type Collection struct {
	Schema string `toml:"schema,omitzero" json:"schema,omitzero"`
	Dir    string `toml:"dir,omitzero"    json:"dir,omitzero"`
}

// Dependency declares a direct external catalog source using the compact inline form:
//
//	shared-infra = { git = "https://...", tag = "v1" }
//	core         = { path = "../core" }
//
// Exactly one source type must be configured. Use Source to obtain the validated variant.
type Dependency struct {
	Git    string `toml:"git,omitzero"    json:"git,omitzero"`
	Tag    string `toml:"tag,omitzero"    json:"tag,omitzero"`
	Branch string `toml:"branch,omitzero" json:"branch,omitzero"`
	Rev    string `toml:"rev,omitzero"    json:"rev,omitzero"`
	Subdir string `toml:"subdir,omitzero" json:"subdir,omitzero"`

	Path string `toml:"path,omitzero" json:"path,omitzero"`
}

// Source is a sealed sum type: GitSource or PathSource.
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
	if d.Path != "" {
		src = PathSource{
			Dir: d.Path,
		}
		names = append(names, "path")
	}

	switch len(names) {
	case 0:
		return nil, errors.New("must specify one of 'git' or 'path'")
	case 1:
		if err := src.validate(); err != nil {
			return nil, err
		}
		return src, nil
	default:
		return nil, fmt.Errorf(
			"multiple sources specified (%s); choose exactly one",
			strings.Join(names, ", "),
		)
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
