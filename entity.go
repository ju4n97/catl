package catl

import (
	"bytes"
	"io/fs"

	"github.com/goccy/go-yaml"

	"github.com/ju4n97/catl/internal/scan"
)

// Entity represents an addressable declarative item within a catalog collection.
// It embeds scan.Entity for filesystem metadata and adds catalog-level resolution and decoding.
type Entity struct {
	scan.Entity

	Namespace   string `json:"namespace"`
	QualifiedID string `json:"qualified_id"`

	cat *Catalog
	raw []byte
}

// Unmarshal decodes the validated manifest content into a user-provided struct.
func (e *Entity) Unmarshal(v any) error {
	return yaml.Unmarshal(e.raw, v)
}

// FS returns a read-only filesystem scoped to the bundle directory.
// For flat entities, calling FS returns an error.
func (e *Entity) FS() (fs.FS, error) {
	return e.Entity.FS(e.cat.fsys)
}

// RawManifest returns the raw unparsed manifest bytes.
func (e *Entity) RawManifest() []byte {
	return bytes.Clone(e.raw)
}
