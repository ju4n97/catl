package scan

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/errdef"
)

// Tree holds all collections and entities discovered within a catalog directory tree.
type Tree struct {
	CatalogName string
	Collections map[string][]Entity
}

// TreeScan scans a catalog directory tree using the provided configuration, enforcing
// clean-checkout rules and discovering all flat and bundle entities.
func TreeScan(fsys fs.FS, cfg *conf.File) (*Tree, error) {
	if cfg == nil {
		return nil, errors.New("cannot scan catalog without configuration")
	}

	if err := validateRoot(fsys, cfg); err != nil {
		return nil, err
	}

	res := &Tree{
		CatalogName: cfg.Catalog.Name,
		Collections: make(map[string][]Entity, len(cfg.Collections)),
	}

	// Discover entities within each declared collection
	for collName, coll := range cfg.Collections {
		collDir := coll.Dir
		if collDir == "" {
			collDir = collName
		}

		entities, err := scanCollection(fsys, collName, collDir)
		if err != nil {
			return nil, err
		}
		res.Collections[collName] = entities
	}

	return res, nil
}

// validateRoot ensures only declared directories, catl files, and ignored dot-files
// exist at the root of the catalog directory tree.
func validateRoot(fsys fs.FS, cfg *conf.File) error {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("read catalog root: %w", err)
	}

	allowedRoots := map[string]struct{}{
		"catl.toml": {},
		"catl.lock": {},
	}

	for collName, coll := range cfg.Collections {
		dir := coll.Dir
		if dir == "" {
			dir = collName
		}
		top, _, _ := strings.Cut(dir, "/")
		allowedRoots[top] = struct{}{}

		if coll.Schema != "" {
			schemaTop, _, _ := strings.Cut(coll.Schema, "/")
			allowedRoots[schemaTop] = struct{}{}
		}
	}

	for _, entry := range entries {
		name := entry.Name()

		if strings.HasPrefix(name, ".") {
			continue
		}

		if err := conf.ValidatePathComponent(name); err != nil {
			return err
		}

		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%w: root entry %q", errdef.ErrForbiddenSymlink, name)
		}

		if _, allowed := allowedRoots[name]; !allowed {
			return fmt.Errorf("%w: %q", errdef.ErrStrayRootEntry, name)
		}
	}

	return nil
}

// scanCollection discovers and parses flat and bundle entities in a collection directory.
func scanCollection(fsys fs.FS, collName, collDir string) ([]Entity, error) {
	entries, err := fs.ReadDir(fsys, collDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// A missing collection directory is treated as an empty collection
			return nil, nil
		}
		return nil, fmt.Errorf("read collection %q: %w", collName, err)
	}

	type candidate struct {
		id    string
		entry fs.DirEntry
	}

	var candidates []candidate
	seenIDs := make(map[string]struct{}, len(entries))

	// Validate path components and detect ID conflicts before walking bundles
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		if err := conf.ValidatePathComponent(name); err != nil {
			return nil, err
		}

		if entry.Type()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("%w: %s/%s", errdef.ErrForbiddenSymlink, collDir, name)
		}

		var id string
		if entry.IsDir() {
			id = name
		} else {
			ext := path.Ext(name)
			if ext != ".yaml" && ext != ".json" {
				return nil, fmt.Errorf("%w: %s/%s", errdef.ErrInvalidManifestType, collDir, name)
			}
			id = strings.TrimSuffix(name, ext)
		}

		if err := conf.ValidateIdent(id); err != nil {
			return nil, fmt.Errorf("collection %q invalid entity ID %q: %w", collName, id, err)
		}

		if _, exists := seenIDs[id]; exists {
			return nil, fmt.Errorf("%w: %q in collection %q", errdef.ErrConflictingEntity, id, collName)
		}
		seenIDs[id] = struct{}{}

		candidates = append(candidates, candidate{id: id, entry: entry})
	}

	// Build flat entities and scan valid bundle directories
	var entities []Entity
	for _, c := range candidates {
		if c.entry.IsDir() {
			bundleEntity, err := scanBundle(fsys, collName, collDir, c.id)
			if err != nil {
				return nil, err
			}
			entities = append(entities, bundleEntity)
		} else {
			entities = append(entities, Entity{
				ID:           c.id,
				Collection:   collName,
				Shape:        EntityShapeFlat,
				ManifestPath: path.Join(collDir, c.entry.Name()),
			})
		}
	}

	slices.SortFunc(entities, func(a, b Entity) int {
		return strings.Compare(a.ID, b.ID)
	})

	return entities, nil
}

// scanBundle validates that a bundle directory contains exactly one manifest and collects companion files.
func scanBundle(fsys fs.FS, collName, collDir, id string) (Entity, error) {
	bundleDir := path.Join(collDir, id)

	var (
		manifestPath   string
		manifestFormat string
		companionFiles []string
	)

	err := fs.WalkDir(fsys, bundleDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		name := d.Name()

		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if err := conf.ValidatePathComponent(name); err != nil {
			return err
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%w: inside bundle %q at %q", errdef.ErrForbiddenSymlink, id, p)
		}

		if d.IsDir() {
			return nil
		}

		relInside := strings.TrimPrefix(p, bundleDir+"/")
		if relInside == "manifest.yaml" || relInside == "manifest.json" {
			ext := path.Ext(name)
			if manifestPath != "" && manifestFormat != ext {
				return fmt.Errorf(
					"%w: %q contains both manifest.yaml and manifest.json",
					errdef.ErrAmbiguousManifest,
					bundleDir,
				)
			}
			manifestPath = p
			manifestFormat = ext
			return nil
		}

		companionFiles = append(companionFiles, relInside)
		return nil
	})
	if err != nil {
		return Entity{}, err
	}

	if manifestPath == "" {
		return Entity{}, fmt.Errorf("%w: %q", errdef.ErrMissingManifest, bundleDir)
	}

	slices.Sort(companionFiles)

	return Entity{
		ID:             id,
		Collection:     collName,
		Shape:          EntityShapeBundle,
		ManifestPath:   manifestPath,
		BundleDir:      bundleDir,
		CompanionFiles: companionFiles,
	}, nil
}
