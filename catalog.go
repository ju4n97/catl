package catl

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/ju4n97/catl/internal/config"
	"github.com/ju4n97/catl/internal/scan"
	"github.com/ju4n97/catl/internal/schema"
)

// Options configures runtime behavior when opening a catalog.
type Options struct {
	// Offline disables network calls and reads exclusively from local vendor directories or cache.
	Offline bool
}

// Catalog represents an immutable, validated catalog loaded into memory.
type Catalog struct {
	fsys      fs.FS
	rootPath  string
	config    *config.File
	tree      *scan.Tree
	entities  map[string]map[string]*Entity // collection -> id -> Entity
	allSorted []*Entity
}

// Open reads and validates a catalog from a directory on the local filesystem.
func Open(ctx context.Context, dir string, opts Options) (*Catalog, error) {
	fsys := os.DirFS(dir)
	cat, err := OpenFS(ctx, fsys, opts)
	if err != nil {
		return nil, err
	}
	cat.rootPath = dir
	return cat, nil
}

// OpenFS reads and validates a catalog from an arbitrary fs.FS filesystem
// (such as testing/fstest.MapFS or embed.FS) without touching the local disk.
func OpenFS(ctx context.Context, fsys fs.FS, opts Options) (*Catalog, error) {
	configData, err := fs.ReadFile(fsys, "catl.toml")
	if err != nil {
		return nil, fmt.Errorf("read catl.toml: %w", err)
	}

	cfg, err := config.Parse(configData)
	if err != nil {
		return nil, err
	}

	tree, err := scan.TreeScan(fsys, cfg)
	if err != nil {
		return nil, err
	}

	validator := schema.NewValidator(fsys)
	var allDiags []schema.Diagnostic

	entityMap := make(map[string]map[string]*Entity, len(cfg.Collections))
	var allSorted []*Entity

	cat := &Catalog{
		fsys:     fsys,
		config:   cfg,
		tree:     tree,
		entities: entityMap,
	}

	for collName, coll := range cfg.Collections {
		scannedEntities := tree.Collections[collName]
		entityMap[collName] = make(map[string]*Entity, len(scannedEntities))

		for _, se := range scannedEntities {
			manifestBytes, err := fs.ReadFile(fsys, se.ManifestPath)
			if err != nil {
				return nil, fmt.Errorf("read manifest %q: %w", se.ManifestPath, err)
			}

			manifest, err := schema.DecodeManifest(manifestBytes, se.ManifestPath)
			if err != nil {
				return nil, err
			}

			var bundleFS fs.FS
			if se.Shape == scan.EntityShapeBundle {
				bundleFS, err = se.FS(fsys)
				if err != nil {
					return nil, err
				}
			}

			if err := validator.ValidateManifest(manifest, coll.Schema, &se, bundleFS); err != nil {
				if valErr, ok := errors.AsType[*schema.ValidationError](err); ok {
					allDiags = append(allDiags, valErr.Diagnostics...)
				} else {
					return nil, err
				}
			}

			ent := &Entity{
				cat:      cat,
				scanItem: se,
				raw:      manifestBytes,
			}
			entityMap[collName][se.ID] = ent
			allSorted = append(allSorted, ent)
		}
	}

	if len(allDiags) > 0 {
		return nil, &schema.ValidationError{Diagnostics: allDiags}
	}

	// Sort entities deterministically: collection asc, then ID asc
	slices.SortFunc(allSorted, func(a, b *Entity) int {
		if c := strings.Compare(a.Collection(), b.Collection()); c != 0 {
			return c
		}
		return strings.Compare(a.ID(), b.ID())
	})
	cat.allSorted = allSorted

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

// Entities returns all entities across collections sorted deterministically.
func (c *Catalog) Entities() []*Entity {
	return c.allSorted
}

// CollectionEntities returns entities belonging to a specific collection, sorted by ID.
func (c *Catalog) CollectionEntities(collection string) ([]*Entity, error) {
	colls, ok := c.entities[collection]
	if !ok {
		return nil, fmt.Errorf("%w: collection %q", ErrNotFound, collection)
	}

	res := make([]*Entity, 0, len(colls))
	for _, ent := range c.allSorted {
		if ent.Collection() == collection {
			res = append(res, ent)
		}
	}
	return res, nil
}

// Resolve retrieves an entity by collection name and qualified identifier (e.g. "app.dark-mode").
// It returns ErrNotQualified if the caller supplies a bare ID without a namespace.
func (c *Catalog) Resolve(collection, qualifiedID string) (*Entity, error) {
	ns, id, hasDot := strings.Cut(qualifiedID, ".")
	if !hasDot {
		return nil, fmt.Errorf("%w: %q (missing namespace prefix)", ErrNotQualified, qualifiedID)
	}

	if ns != c.Name() {
		// TODO: external dependency namespaces will be resolved here
		return nil, fmt.Errorf("%w: namespace %q", ErrNotFound, ns)
	}

	collMap, ok := c.entities[collection]
	if !ok {
		return nil, fmt.Errorf("%w: collection %q", ErrNotFound, collection)
	}

	ent, ok := collMap[id]
	if !ok {
		return nil, fmt.Errorf("%w: entity %q in collection %q", ErrNotFound, id, collection)
	}

	return ent, nil
}
