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

// Entity represents a discovered catalog item within a collection.
type Entity struct {
	ID             string      `json:"id"`
	Collection     string      `json:"collection"`
	ManifestPath   string      `json:"path"`
	BundleDir      string      `json:"-"`
	CompanionFiles []string    `json:"companion_files,omitzero"`
	Shape          EntityShape `json:"-"`
}

// IsFlat reports whether the entity is a single manifest file.
func (e *Entity) IsFlat() bool {
	return e.Shape == EntityShapeFlat
}

// IsBundle reports whether the entity is a bundle directory with companion files.
func (e *Entity) IsBundle() bool {
	return e.Shape == EntityShapeBundle
}

// FS returns a read-only filesystem scoped to the bundle directory.
// Returns an error if called on a flat entity.
func (e *Entity) FS(fsys fs.FS) (fs.FS, error) {
	if !e.IsBundle() {
		return nil, errors.New("cannot create bundle filesystem for flat entity")
	}

	sub, err := fs.Sub(fsys, e.BundleDir)
	if err != nil {
		return nil, fmt.Errorf("open bundle fs for %q: %w", e.ID, err)
	}

	return sub, nil
}
