package scan

import (
	"errors"
	"fmt"
	"io/fs"
)

// EntityShape describes the filesystem layout of an entity: a single flat manifest
// or a bundle directory containing a manifest and companion files.
type EntityShape int

const (
	// EntityShapeFlat represents a single manifest file placed directly in a collection directory.
	EntityShapeFlat EntityShape = iota

	// EntityShapeBundle represents an entity directory containing a manifest and optional companion files.
	EntityShapeBundle
)

// String returns a human-readable description of the shape.
func (s EntityShape) String() string {
	switch s {
	case EntityShapeFlat:
		return "flat"
	case EntityShapeBundle:
		return "bundle"
	default:
		return "unknown"
	}
}

// Entity represents a discovered catalog item within a collection.
type Entity struct {
	ID             string
	Collection     string
	Shape          EntityShape
	ManifestPath   string   // Path relative to catalog root (e.g. "services/auth.yaml" or "services/billing/manifest.yaml")
	BundleDir      string   // Directory relative to catalog root (empty for flat entities)
	CompanionFiles []string // Relative paths inside the bundle directory (excluding manifest and dotfiles)
}

// FS returns a read-only filesystem scoped to the bundle directory.
// For flat entities, calling FS returns an error because flat entities can't have companion files.
func (e *Entity) FS(fsys fs.FS) (fs.FS, error) {
	if e.Shape != EntityShapeBundle {
		return nil, errors.New("cannot create bundle filesystem for flat entity")
	}

	sub, err := fs.Sub(fsys, e.BundleDir)
	if err != nil {
		return nil, fmt.Errorf("open bundle fs for %q: %w", e.ID, err)
	}

	return sub, nil
}
