// Package conf loads and validates catl.toml, the root configuration of a
// catalog. It models the catalog identity, declared collections, and direct
// dependencies, and enforces identifier, path, and source invariants.
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

// Validate checks catalog identity, collection paths, and dependency sources.
func (f *File) Validate() error {
	if err := ValidateIdent(f.Catalog.Name); err != nil {
		return fmt.Errorf("catalog name: %w", err)
	}
	if len(f.Collections) == 0 {
		return errors.New("at least one collection is required")
	}
	for name := range f.Collections {
		if err := ValidateIdent(name); err != nil {
			return fmt.Errorf("invalid collection identifier %q: %w", name, err)
		}
	}
	if err := f.ValidateCollections(); err != nil {
		return err
	}
	for name, dep := range f.Dependencies {
		if err := ValidateIdent(name); err != nil {
			return fmt.Errorf("invalid dependency identifier %q: %w", name, err)
		}
		if _, err := dep.Source(); err != nil {
			return fmt.Errorf("dependency %q: %w", name, err)
		}
	}
	return nil
}

// ValidateCollections checks relative path invariants and prevents collection storage directories from overlapping.
func (f *File) ValidateCollections() error {
	seenDirs := make(map[string]string, len(f.Collections))

	for name, coll := range f.Collections {
		if coll.Schema != "" {
			coll.Schema = strings.TrimPrefix(coll.Schema, "./")
			if err := ValidateRelativePath(coll.Schema); err != nil {
				return fmt.Errorf("collection %q schema: %w", name, err)
			}
		}

		dir := coll.Dir
		if dir == "" {
			dir = name
		} else {
			coll.Dir = strings.TrimPrefix(coll.Dir, "./")
			dir = coll.Dir
			if err := ValidateRelativePath(dir); err != nil {
				return fmt.Errorf("collection %q dir: %w", name, err)
			}
		}

		f.Collections[name] = coll
		for existingDir, existingName := range seenDirs {
			if dir == existingDir || strings.HasPrefix(dir, existingDir+"/") || strings.HasPrefix(existingDir, dir+"/") {
				return fmt.Errorf(
					"collections %q and %q have overlapping directories %q and %q",
					existingName,
					name,
					existingDir,
					dir,
				)
			}
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
	subdir := strings.TrimPrefix(s.Subdir, "./")
	if s.Repo == "" {
		return errors.New("git: 'repo' is required")
	}
	if !strings.HasPrefix(s.Repo, "https://") && !strings.HasPrefix(s.Repo, "ssh://") {
		return errors.New("git: 'repo' must use https:// or ssh://")
	}
	if s.Repo == "https://" || s.Repo == "ssh://" {
		return errors.New("git: 'repo' must contain a repository after the scheme")
	}
	if countSet(s.Tag, s.Branch, s.Rev) != 1 {
		return errors.New("git: exactly one of 'tag', 'branch', or 'rev' is required")
	}
	if s.Rev != "" {
		if len(s.Rev) != 40 {
			return errors.New("git: 'rev' must be a 40-character lowercase hexadecimal commit hash")
		}
		for _, c := range s.Rev {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return errors.New("git: 'rev' must be a 40-character lowercase hexadecimal commit hash")
			}
		}
	}
	if subdir != "" {
		if err := ValidateRelativePath(subdir); err != nil {
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
		subdir := strings.TrimPrefix(d.Subdir, "./")
		if d.Subdir != "" && subdir == "" {
			return nil, fmt.Errorf("git: invalid subdir: %w", ValidateRelativePath(subdir))
		}
		src = GitSource{
			Repo:   d.Git,
			Tag:    d.Tag,
			Branch: d.Branch,
			Rev:    d.Rev,
			Subdir: subdir,
		}
		names = append(names, "git")
	}
	if d.Path != "" {
		if d.Git != "" || d.Tag != "" || d.Branch != "" || d.Rev != "" || d.Subdir != "" {
			return nil, errors.New("path: Git-only fields cannot be set")
		}
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

// countSet counts how many strings in the slice are non-empty.
func countSet(ss ...string) (n int) {
	for _, s := range ss {
		if s != "" {
			n++
		}
	}
	return n
}
