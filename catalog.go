// Package catl loads and validates declarative entity catalogs.
//
// A catalog is a tree of YAML or JSON manifests organized into collections,
// declared by catl.toml and pinned by catl.lock. Open and OpenFS load a
// catalog into an immutable Catalog that is safe for concurrent use;
// Registry holds several catalogs side by side. Validation failures are
// reported as ValidationError values. Loading performs no network I/O:
// dependencies come from local paths, .catl/vendor, or the global cache.
package catl

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/hash"
	"github.com/ju4n97/catl/internal/lock"
	"github.com/ju4n97/catl/internal/manifest"
	"github.com/ju4n97/catl/internal/scan"
	"github.com/ju4n97/catl/internal/source"
)

// Catalog is an immutable, validated catalog loaded into memory.
// It's safe for concurrent use across multiple goroutines.
type Catalog struct {
	fsys     fs.FS
	rootPath string
	config   *conf.File

	// Local collection names, sorted.
	collectionNames []string

	// Local entities grouped by collection, sorted by entity ID.
	// Used by Entities() to avoid rebuilding the list on every call.
	entitiesByCollection map[string][]*Entity

	// collection -> entity ID -> entity. Used for direct lookups in Resolve().
	entityByCollectionAndID map[string]map[string]*Entity

	// All local entities, sorted by collection and then ID.
	localEntities []*Entity

	// Loaded direct dependencies, keyed by namespace.
	dependenciesByNamespace map[string]*Catalog

	// Relative paths to dependency roots, used when returning their manifest paths.
	// For example, "core" might point to "../core" or ".catl/vendor/core".
	dependencyPathPrefixes map[string]string
}

// Open loads a catalog from dir, validates its manifests, and builds the lookup
// indexes. It also resolves direct dependencies from local paths, .catl/vendor,
// or the global cache ($CATL_CACHE_DIR).
//
// If .catl/vendor exists, remote dependencies must be there. Open won't fall
// back to the global cache, and returns ErrMissing if a dependency is missing.
//
// A missing or outdated catl.lock causes an error wrapping ErrLockStale.
// Open won't fetch anything from the network or update the lockfile for you.
//
// Only direct dependencies are loaded. Dependencies declared by those catalogs
// are ignored.
//
// The returned catalog is read-only and safe for concurrent use.
func Open(ctx context.Context, dir string) (*Catalog, error) {
	cat, err := openInternal(ctx, os.DirFS(dir), dir, false)
	if err != nil {
		return nil, err
	}
	cat.rootPath = dir
	return cat, nil
}

// OpenFS loads a catalog from an fs.FS, like embed.FS or fstest.MapFS.
//
// Since fsys is separate from the host filesystem, OpenFS can't resolve local
// sibling paths or use the global cache. Remote dependencies must already be
// vendored under .catl/vendor/<namespace>, and catl.lock must be present.
//
// If a dependency is missing from vendor, or its tree hash doesn't match the
// lockfile, OpenFS returns ErrMissing or ErrTamperedVendor.
//
// Like Open, the returned catalog is safe for concurrent use.
func OpenFS(ctx context.Context, fsys fs.FS) (*Catalog, error) {
	return openInternal(ctx, fsys, "", false)
}

// openInternal reads the config, scans the catalog, validates manifests, and
// builds the lookup indexes. It can also load direct dependencies, unless this
// catalog is being opened as a dependency itself.
func openInternal(ctx context.Context, fsys fs.FS, rootPath string, isDependency bool) (*Catalog, error) {
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
	var allDiags []Diagnostic

	collNames := make([]string, 0, len(cfg.Collections))
	for name := range cfg.Collections {
		collNames = append(collNames, name)
	}
	slices.Sort(collNames)

	cat := &Catalog{
		fsys:                    fsys,
		config:                  cfg,
		collectionNames:         collNames,
		entitiesByCollection:    make(map[string][]*Entity, len(collNames)),
		entityByCollectionAndID: make(map[string]map[string]*Entity, len(collNames)),
		dependenciesByNamespace: make(map[string]*Catalog),
		dependencyPathPrefixes:  make(map[string]string),
	}

	for _, collName := range collNames {
		coll := cfg.Collections[collName]
		scanned := tree.Collections[collName]

		cat.entitiesByCollection[collName] = make([]*Entity, 0, len(scanned))
		cat.entityByCollectionAndID[collName] = make(map[string]*Entity, len(scanned))

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
				if valErr, ok := errors.AsType[*ValidationError](err); ok {
					allDiags = append(allDiags, valErr.Diagnostics...)
				} else {
					return nil, err
				}
			}

			ent := newEntity(cat, cfg.Catalog.Name, se, manifestBytes)

			cat.entitiesByCollection[collName] = append(cat.entitiesByCollection[collName], ent)
			cat.entityByCollectionAndID[collName][se.ID] = ent
			cat.localEntities = append(cat.localEntities, ent)
		}
	}

	if len(allDiags) > 0 {
		return nil, &ValidationError{Diagnostics: allDiags}
	}

	// Dependencies are direct only. This doesn't load dependencies of dependencies
	if isDependency || len(cfg.Dependencies) == 0 {
		return cat, nil
	}

	if err := cat.loadDependencies(ctx, rootPath); err != nil {
		return nil, err
	}

	return cat, nil
}

// loadDependencies loads this catalog's direct dependencies and records the
// path prefixes needed to return their manifest paths relative to this catalog.
func (c *Catalog) loadDependencies(ctx context.Context, rootPath string) error {
	if rootPath != "" {
		// Disk-backed catalog: resolve local paths, vendor entries, and the global cache
		lockPath := filepath.Join(rootPath, "catl.lock")
		var lockFile *lock.File
		if _, err := os.Stat(lockPath); err == nil {
			var err error
			lockFile, err = lock.Load(lockPath)
			if err != nil {
				return err
			}
		}

		resolver := source.NewResolver(rootPath)
		resolvedMap, err := resolver.ResolveAll(c.config.Dependencies, lockFile)
		if err != nil {
			return err
		}

		for name, resolved := range resolvedMap {
			depCat, err := openInternal(ctx, os.DirFS(resolved.Path), resolved.Path, true)
			if err != nil {
				return fmt.Errorf("open dependency %q: %w", name, err)
			}
			depCat.rootPath = resolved.Path
			c.dependenciesByNamespace[name] = depCat

			absRoot, err := filepath.Abs(rootPath)
			if err != nil {
				absRoot = rootPath
			}

			if rel, err := filepath.Rel(absRoot, resolved.Path); err == nil {
				c.dependencyPathPrefixes[name] = filepath.ToSlash(rel)
			} else {
				c.dependencyPathPrefixes[name] = filepath.ToSlash(resolved.Path)
			}
		}
		return nil
	}

	// OpenFS only has access to the filesystem passed by the caller
	var lockFile *lock.File
	if lockBytes, err := fs.ReadFile(c.fsys, "catl.lock"); err == nil {
		var err error
		lockFile, err = lock.Parse(lockBytes)
		if err != nil {
			return err
		}
	}

	if err := lock.VerifyFreshness(c.config.Dependencies, lockFile); err != nil {
		return err
	}

	for name, dep := range c.config.Dependencies {
		if dep.Path != "" {
			return fmt.Errorf("local path dependency %q not supported in pure in-memory filesystem", name)
		}

		req, err := lock.CanonicalRequested(dep)
		if err != nil {
			return err
		}

		entry, ok := lockFile.FindEntry(req)
		if !ok {
			return fmt.Errorf("%w: missing pin for %q", ErrLockStale, name)
		}

		vendorPath := path.Join(".catl", "vendor", name)
		sub, err := fs.Sub(c.fsys, vendorPath)
		if err != nil {
			return fmt.Errorf("%w: dependency %q not found in vendor: %w", ErrMissing, name, err)
		}

		actualHash, err := hash.Tree(sub)
		if err != nil {
			return fmt.Errorf("tree hash for vendored %q: %w", name, err)
		}
		if actualHash != entry.Tree {
			return fmt.Errorf("%w: dependency %q tree hash is %s, want %s", ErrTamperedVendor, name, actualHash, entry.Tree)
		}

		depCat, err := openInternal(ctx, sub, "", true)
		if err != nil {
			return fmt.Errorf("open vendored dependency %q: %w", name, err)
		}
		c.dependenciesByNamespace[name] = depCat
		c.dependencyPathPrefixes[name] = vendorPath
	}

	return nil
}

// Name returns the catalog namespace from catl.toml.
func (c *Catalog) Name() string {
	return c.config.Catalog.Name
}

// RootPath returns the directory used to open the catalog. It's empty for
// catalogs opened with OpenFS.
func (c *Catalog) RootPath() string {
	return c.rootPath
}

// FS returns the filesystem backing this catalog.
func (c *Catalog) FS() fs.FS {
	return c.fsys
}

// TreeHash returns the deterministic SHA-256 hash of the catalog's filesystem tree.
func (c *Catalog) TreeHash() (string, error) {
	return hash.Tree(c.fsys)
}

// Collections returns the local collection names sorted alphabetically.
func (c *Catalog) Collections() []string {
	return slices.Clone(c.collectionNames)
}

// Dependencies returns the loaded direct dependency namespaces sorted alphabetically.
func (c *Catalog) Dependencies() []string {
	deps := make([]string, 0, len(c.dependenciesByNamespace))
	for name := range c.dependenciesByNamespace {
		deps = append(deps, name)
	}
	slices.Sort(deps)
	return deps
}

// Dependency returns a loaded dependency by namespace. The bool is false if
// no dependency with that name was loaded.
func (c *Catalog) Dependency(name string) (*Catalog, bool) {
	dep, ok := c.dependenciesByNamespace[name]
	return dep, ok
}

// AllCollections returns collection names from this catalog and its direct
// dependencies. Duplicate names are removed and the result is sorted.
func (c *Catalog) AllCollections() []string {
	seen := make(map[string]struct{}, len(c.collectionNames))
	for _, name := range c.collectionNames {
		seen[name] = struct{}{}
	}
	for _, dep := range c.dependenciesByNamespace {
		for _, name := range dep.Collections() {
			seen[name] = struct{}{}
		}
	}

	res := make([]string, 0, len(seen))
	for name := range seen {
		res = append(res, name)
	}
	slices.Sort(res)
	return res
}

// Schema returns the raw JSON Schema bytes configured for a local collection.
//
// If the collection has no schema, it returns nil, nil. In that case, manifests
// are only checked for valid YAML/JSON syntax.
//
// Returns an error wrapping ErrNotFound if the collection isn't declared here.
// To get a dependency's schema, first retrieve that catalog with Dependency().
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

// Entities returns local entities, without anything from dependencies.
//
// With no filters, the result is sorted by collection name and then entity ID.
// Pass collection names to limit the result to those collections.
//
// The returned slice and every returned *Entity are independent copies, safe
// for the caller to mutate. Use AllEntities() to include dependencies.
func (c *Catalog) Entities(collections ...string) []*Entity {
	if len(collections) == 0 {
		return cloneEntities(c.localEntities)
	}

	var res []*Entity
	for _, coll := range collections {
		res = append(res, cloneEntities(c.entitiesByCollection[coll])...)
	}
	return res
}

// AllEntities returns entities from this catalog and its direct dependencies.
//
// You can pass collection names to filter the result, including collections
// that only exist in a dependency. Results are sorted by collection name,
// then by qualified ID (for example, "app.api" before "core.postgres").
//
// Manifest paths for dependency entities are prefixed with the path to the
// dependency relative to this catalog's root.
//
// The slice is newly allocated and all returned entities are copies, safe for
// the caller to mutate.
func (c *Catalog) AllEntities(collections ...string) []*Entity {
	var all []*Entity
	all = append(all, c.Entities(collections...)...)

	for _, depName := range c.Dependencies() {
		depCat := c.dependenciesByNamespace[depName]
		prefix := c.dependencyPathPrefixes[depName]

		for _, ent := range depCat.Entities(collections...) {
			entCopy := ent.clone()
			if prefix != "" {
				entCopy.ManifestPath = path.Join(prefix, ent.ManifestPath)
			}
			all = append(all, entCopy)
		}
	}

	slices.SortFunc(all, func(a, b *Entity) int {
		if cmp := strings.Compare(a.Collection, b.Collection); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.QualifiedID, b.QualifiedID)
	})

	return all
}

// Resolve looks up an entity by collection and qualified ID, such as
// "platform.billing" or "core.postgres". IDs without a namespace prefix
// return an error wrapping ErrNotQualified.
//
// If the namespace is this catalog's name, it looks locally. Otherwise, it
// checks the loaded direct dependencies. It doesn't search transitively.
//
// For dependency entities, ManifestPath is adjusted to be relative to this
// catalog's root. That way, callers can pass the path to shell tools or editors
// without having to work out the relative path themselves.
//
// This doesn't change companion file access: Entity.FS() still reads from
// the entity's own bundle.
//
// Returns an error wrapping ErrNotFound if the collection, namespace, or entity
// can't be found.
func (c *Catalog) Resolve(collection, qualifiedID string) (*Entity, error) {
	ns, id, hasDot := strings.Cut(qualifiedID, ".")
	if !hasDot {
		return nil, fmt.Errorf("%w: %q (missing namespace prefix)", ErrNotQualified, qualifiedID)
	}

	// Look in this catalog first
	if ns == c.Name() {
		byID, ok := c.entityByCollectionAndID[collection]
		if !ok {
			return nil, fmt.Errorf("%w: collection %q", ErrNotFound, collection)
		}
		ent, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("%w: entity %q in collection %q", ErrNotFound, id, collection)
		}
		return ent.clone(), nil
	}

	// Otherwise, try a direct dependency
	depCat, ok := c.dependenciesByNamespace[ns]
	if !ok {
		return nil, fmt.Errorf("%w: namespace %q", ErrNotFound, ns)
	}

	ent, err := depCat.Resolve(collection, qualifiedID)
	if err != nil {
		return nil, err
	}

	// Make the path relative to the current catalog's root
	if prefix := c.dependencyPathPrefixes[ns]; prefix != "" {
		entCopy := ent.clone()
		entCopy.ManifestPath = path.Join(prefix, ent.ManifestPath)
		return entCopy, nil
	}

	return ent, nil
}

// ValidateManifest checks YAML or JSON content against the schema configured
// for the given collection. It works on the provided bytes, without writing
// anything to disk.
//
// Content is parsed using YAML 1.2 core schema rules, then checked against
// JSON Schema Draft 2020-12.
//
// If the collection has no schema, only parsing is done. An invalid manifest
// returns an error wrapping *ValidationError with line-by-line diagnostics.
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
