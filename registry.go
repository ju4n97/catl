package catl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
)

// Registry manages multiple isolated tenant catalogs loaded concurrently in memory.
type Registry struct {
	dir      string
	opts     Options
	catalogs atomic.Pointer[map[string]*Catalog]
}

// NewRegistry wraps pre-loaded catalogs into an in-memory Registry.
func NewRegistry(catalogs ...*Catalog) *Registry {
	reg := &Registry{}
	catMap := make(map[string]*Catalog, len(catalogs))
	for _, c := range catalogs {
		catMap[c.Name()] = c
	}
	reg.catalogs.Store(&catMap)
	return reg
}

// OpenRegistry discovers and loads all catalog subdirectories within rootDir.
func OpenRegistry(ctx context.Context, rootDir string, opts Options) (*Registry, error) {
	reg := &Registry{
		dir:  rootDir,
		opts: opts,
	}

	if err := reg.Reload(ctx); err != nil {
		return nil, err
	}

	return reg, nil
}

// Dir returns the root filesystem path of the registry, or empty if created via NewRegistry.
func (r *Registry) Dir() string {
	return r.dir
}

// Names returns a sorted list of all active catalog namespaces in the registry.
func (r *Registry) Names() []string {
	cats := *r.catalogs.Load()
	names := make([]string, 0, len(cats))
	for name := range cats {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Catalogs returns all loaded catalogs sorted by namespace name.
func (r *Registry) Catalogs() []*Catalog {
	names := r.Names()
	cats := *r.catalogs.Load()
	res := make([]*Catalog, 0, len(names))
	for _, name := range names {
		res = append(res, cats[name])
	}
	return res
}

// Catalog retrieves an individual catalog by its namespace.
func (r *Registry) Catalog(name string) (*Catalog, bool) {
	cats := *r.catalogs.Load()
	cat, ok := cats[name]
	return cat, ok
}

// Reload rescans the registry directory on disk and atomically updates all catalogs in memory.
func (r *Registry) Reload(ctx context.Context) error {
	if r.dir == "" {
		return errors.New("cannot reload in-memory registry without backing directory")
	}

	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return fmt.Errorf("read registry dir: %w", err)
	}

	loaded := make(map[string]*Catalog)

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		catDir := filepath.Join(r.dir, entry.Name())
		if _, err := os.Stat(filepath.Join(catDir, "catl.toml")); err == nil {
			cat, err := Open(ctx, catDir, r.opts)
			if err != nil {
				return fmt.Errorf("load catalog %q: %w", entry.Name(), err)
			}
			loaded[cat.Name()] = cat
		}
	}

	r.catalogs.Store(&loaded)
	return nil
}
