package catl

import (
	"context"
	"io/fs"
	"slices"

	"github.com/goccy/go-yaml"

	"github.com/ju4n97/catl/internal/scan"
)

// Entity represents an addressable declarative item within a catalog collection.
type Entity struct {
	cat      *Catalog
	scanItem scan.Entity
	raw      []byte
}

// Collection returns the collection name the entity belongs to (e.g. "flags").
func (e *Entity) Collection() string {
	return e.scanItem.Collection
}

// Namespace returns the namespace owning the entity (the catalog name).
func (e *Entity) Namespace() string {
	return e.cat.Name()
}

// ID returns the bare identifier of the entity (e.g. "dark-mode").
func (e *Entity) ID() string {
	return e.scanItem.ID
}

// QualifiedID returns the canonical address of the entity: "namespace.id".
func (e *Entity) QualifiedID() string {
	return e.Namespace() + "." + e.ID()
}

// ManifestPath returns the relative path from the catalog root to the manifest file.
func (e *Entity) ManifestPath() string {
	return e.scanItem.ManifestPath
}

// Shape returns whether the entity is a flat single file or a bundle directory.
func (e *Entity) Shape() scan.EntityShape {
	return e.scanItem.Shape
}

// CompanionFiles returns a slice of paths to companion files relative to the bundle directory.
// Returns an empty slice for flat entities.
func (e *Entity) CompanionFiles() []string {
	return slices.Clone(e.scanItem.CompanionFiles)
}

// Unmarshal decodes the validated manifest content into a user-provided struct.
func (e *Entity) Unmarshal(v any) error {
	return yaml.Unmarshal(e.raw, v)
}

// FS returns a read-only filesystem scoped to the bundle directory.
// It returns an error if called on a flat entity.
func (e *Entity) FS(ctx context.Context) (fs.FS, error) {
	return e.scanItem.FS(e.cat.fsys)
}

// RawManifest returns the raw unparsed bytes of the entity manifest.
func (e *Entity) RawManifest() []byte {
	return e.raw
}
