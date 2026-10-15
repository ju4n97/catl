// Package catl provides a catalog manager and validator for declarative data stored in filesystem directories.
package catl

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/hash"
	"github.com/ju4n97/catl/internal/manifest"
	"github.com/ju4n97/catl/internal/scan"
)

// Catalog represents an immutable, validated catalog loaded into memory.
type Catalog struct {
	fsys        fs.FS
	rootPath    string
	config      *conf.File
	collections []string                      // Sorted declared collection names
	byColl      map[string][]*Entity          // collection -> entities sorted by ID
	byCollID    map[string]map[string]*Entity // collection -> ID -> entity
	allSorted   []*Entity                     // All entities sorted by collection, then ID
}

// Open reads and validates a catalog from a local filesystem directory.
func Open(ctx context.Context, dir string) (*Catalog, error) {
	cat, err := OpenFS(ctx, os.DirFS(dir))
	if err != nil {
		return nil, err
	}
	cat.rootPath = dir
	return cat, nil
}

// OpenFS reads and validates a catalog from an arbitrary fs.FS filesystem
// (such as testing/fstest.MapFS or embed.FS) without accessing the host disk.
func OpenFS(ctx context.Context, fsys fs.FS) (*Catalog, error) {
	configData, err := fs.ReadFile(fsys, "catl.toml")
	if err != nil {
		return nil, fmt.Errorf("read catl.toml: %w", err)
	}

	cfg, err := conf.Parse(configData)
	if err != nil {
		return nil, err
	}

	tree, err := scan.TreeScan(fsys, cfg)
	if err != nil {
		return nil, err
	}

	validator := manifest.NewValidator(fsys)
	var allDiags []manifest.Diagnostic

	collNames := make([]string, 0, len(cfg.Collections))
	for name := range cfg.Collections {
		collNames = append(collNames, name)
	}
	slices.Sort(collNames)

	cat := &Catalog{
		fsys:        fsys,
		config:      cfg,
		collections: collNames,
		byColl:      make(map[string][]*Entity, len(collNames)),
		byCollID:    make(map[string]map[string]*Entity, len(collNames)),
	}

	for _, collName := range collNames {
		coll := cfg.Collections[collName]
		scanned := tree.Collections[collName]

		cat.byColl[collName] = make([]*Entity, 0, len(scanned))
		cat.byCollID[collName] = make(map[string]*Entity, len(scanned))

		for _, se := range scanned {
			manifestBytes, err := fs.ReadFile(fsys, se.ManifestPath)
			if err != nil {
				return nil, fmt.Errorf("read manifest %q: %w", se.ManifestPath, err)
			}

			decoded, err := manifest.Decode(manifestBytes, se.ManifestPath)
			if err != nil {
				return nil, err
			}

			var bundleFS fs.FS
			if se.IsBundle() {
				bundleFS, err = se.FS(fsys)
				if err != nil {
					return nil, err
				}
			}

			if err := validator.Validate(decoded, coll.Schema, &se, bundleFS); err != nil {
				if valErr, ok := errors.AsType[*manifest.ValidationError](err); ok {
					allDiags = append(allDiags, valErr.Diagnostics...)
				} else {
					return nil, err
				}
			}

			ent := &Entity{
				Entity:      se,
				Namespace:   cfg.Catalog.Name,
				QualifiedID: cfg.Catalog.Name + "." + se.ID,
				cat:         cat,
				raw:         manifestBytes,
			}

			cat.byColl[collName] = append(cat.byColl[collName], ent)
			cat.byCollID[collName][se.ID] = ent
			cat.allSorted = append(cat.allSorted, ent)
		}
	}

	if len(allDiags) > 0 {
		return nil, &manifest.ValidationError{Diagnostics: allDiags}
	}

	return cat, nil
}

// Name returns the catalog's primary namespace as declared in catl.toml.
func (c *Catalog) Name() string {
	return c.config.Catalog.Name
}

// RootPath returns the local directory path if opened from disk, or empty if opened via OpenFS.
func (c *Catalog) RootPath() string {
	return c.rootPath
}

// FS returns the underlying read-only filesystem backing the catalog.
func (c *Catalog) FS() fs.FS {
	return c.fsys
}

// TreeHash returns the deterministic SHA-256 tree hash identifying catalog contents.
func (c *Catalog) TreeHash() (string, error) {
	return hash.Tree(c.fsys)
}

// Collections returns all declared collection names sorted in alphabetical order.
func (c *Catalog) Collections() []string {
	return slices.Clone(c.collections)
}

// Schema returns the raw JSON Schema bytes for a declared collection.
// If the collection exists but has no schema configured (permissive mode),
// it returns nil, nil. If the collection does not exist, it returns an error wrapping ErrNotFound.
func (c *Catalog) Schema(collection string) ([]byte, error) {
	coll, ok := c.config.Collections[collection]
	if !ok {
		return nil, fmt.Errorf("%w: collection %q", ErrNotFound, collection)
	}
	if coll.Schema == "" {
		return nil, nil
	}
	return fs.ReadFile(c.fsys, coll.Schema)
}

// Entities returns all entities in deterministic order. If collection names are provided,
// only entities belonging to those collections are returned.
func (c *Catalog) Entities(collections ...string) []*Entity {
	if len(collections) == 0 {
		return slices.Clone(c.allSorted)
	}

	var res []*Entity
	for _, coll := range collections {
		res = append(res, c.byColl[coll]...)
	}
	return res
}

// Resolve retrieves an entity by collection name and qualified identifier.
// Bare identifiers without a namespace return ErrNotQualified.
func (c *Catalog) Resolve(collection, qualifiedID string) (*Entity, error) {
	ns, id, hasDot := strings.Cut(qualifiedID, ".")
	if !hasDot {
		return nil, fmt.Errorf("%w: %q (missing namespace prefix)", ErrNotQualified, qualifiedID)
	}

	if ns != c.Name() {
		return nil, fmt.Errorf("%w: namespace %q", ErrNotFound, ns)
	}

	byID, ok := c.byCollID[collection]
	if !ok {
		return nil, fmt.Errorf("%w: collection %q", ErrNotFound, collection)
	}

	ent, ok := byID[id]
	if !ok {
		return nil, fmt.Errorf("%w: entity %q in collection %q", ErrNotFound, id, collection)
	}

	return ent, nil
}

// ValidateManifest parses and validates an uncommitted YAML or JSON manifest string in-memory
// against a collection's configured schema without touching the filesystem.
func (c *Catalog) ValidateManifest(collection string, content []byte) error {
	coll, ok := c.config.Collections[collection]
	if !ok {
		return fmt.Errorf("%w: collection %q", ErrNotFound, collection)
	}

	m, err := manifest.Decode(content, collection+"/manifest.yaml")
	if err != nil {
		return err
	}

	if coll.Schema == "" {
		return nil
	}

	return manifest.NewValidator(c.fsys).Validate(m, coll.Schema, nil, nil)
}
