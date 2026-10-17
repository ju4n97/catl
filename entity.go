package catl

import (
	"bytes"
	"io/fs"
	"slices"

	"github.com/goccy/go-yaml"

	"github.com/ju4n97/catl/internal/scan"
)

// Entity represents an addressable declarative item within a catalog collection.
// The exported fields are the public projection used for JSON output. Behavior
// such as bundle detection and companion file access is delegated to the
// scanner's view held in scan, which is the source of truth for entity shape.
type Entity struct {
	ID             string   `json:"id"`
	Collection     string   `json:"collection"`
	Namespace      string   `json:"namespace"`
	QualifiedID    string   `json:"qualified_id"`
	ManifestPath   string   `json:"path"`
	CompanionFiles []string `json:"companion_files,omitzero"`

	scan scan.Entity
	cat  *Catalog
	raw  []byte
}

// newEntity projects a scanned entity into the public Entity type.
func newEntity(cat *Catalog, namespace string, se scan.Entity, raw []byte) *Entity {
	return &Entity{
		ID:             se.ID,
		Collection:     se.Collection,
		Namespace:      namespace,
		QualifiedID:    namespace + "." + se.ID,
		ManifestPath:   se.ManifestPath,
		CompanionFiles: slices.Clone(se.CompanionFiles),
		scan:           se,
		cat:            cat,
		raw:            raw,
	}
}

// Unmarshal decodes the validated manifest content into v. Unknown fields are
// allowed so callers can decode partial views of a manifest.
func (e *Entity) Unmarshal(v any) error {
	return yaml.Unmarshal(e.raw, v)
}

// IsFlat reports whether the entity is a single manifest file.
func (e *Entity) IsFlat() bool {
	return e.scan.IsFlat()
}

// IsBundle reports whether the entity is a bundle directory with companion files.
func (e *Entity) IsBundle() bool {
	return e.scan.IsBundle()
}

// FS returns a read-only filesystem scoped to the bundle directory.
// For flat entities, calling FS returns an error.
func (e *Entity) FS() (fs.FS, error) {
	return e.scan.FS(e.cat.fsys)
}

// RawManifest returns the raw unparsed manifest bytes.
func (e *Entity) RawManifest() []byte {
	return bytes.Clone(e.raw)
}

// clone returns a copy of e whose CompanionFiles slice is independent of the
// catalog's internal state.
func (e *Entity) clone() *Entity {
	c := *e
	c.CompanionFiles = slices.Clone(e.CompanionFiles)
	return &c
}

// cloneEntities returns independent copies of ents.
func cloneEntities(ents []*Entity) []*Entity {
	out := make([]*Entity, 0, len(ents))
	for _, ent := range ents {
		out = append(out, ent.clone())
	}
	return out
}
