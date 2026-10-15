package source

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/hash"
	"github.com/ju4n97/catl/internal/lock"
)

var (
	// ErrMissing indicates a dependency's catalog files were not found on disk.
	ErrMissing = errors.New("dependency files not found")

	// ErrTamperedVendor indicates vendored dependency content does not match the recorded tree hash.
	ErrTamperedVendor = errors.New("vendor directory content does not match tree hash")

	// ErrCorruptCache indicates cached dependency content does not match the recorded tree hash.
	ErrCorruptCache = errors.New("cached dependency content does not match tree hash")

	// ErrTargetMismatch indicates the resolved catalog's declared name does not match the dependency key.
	ErrTargetMismatch = errors.New("target catalog name does not match dependency key")

	// ErrLockRequired indicates a remote dependency is missing its required lockfile pin.
	ErrLockRequired = errors.New("remote dependency requires lockfile entry")
)

// Resolved represents an external catalog dependency resolved to an on-disk location.
type Resolved struct {
	Name       string // Dependency namespace matching the key in catl.toml
	Path       string // Absolute filesystem path to the dependency's root directory
	FS         fs.FS  // Read-only filesystem scoped to the dependency catalog
	IsVendored bool   // True if resolved from the local .catl/vendor directory
	IsPath     bool   // True if resolved from a local path source
	TreeHash   string // Recorded SHA-256 tree hash from catl.lock (empty for path sources)
}

// TreeKey generates a deterministic, collision-proof cache key for a Git commit and optional subdirectory.
// For root checkouts (subdir empty or "."), it returns the 40-character commit SHA.
// For subdirectories, it appends a 16-character hex digest of the normalized subdirectory path.
func TreeKey(commit, subdir string) string {
	cleaned := path.Clean(strings.Trim(filepath.ToSlash(subdir), "/"))
	if cleaned == "" || cleaned == "." {
		return commit
	}

	h := sha256.Sum256([]byte(cleaned))
	return fmt.Sprintf("%s-%x", commit, h[:8])
}

// DefaultCacheDir returns the authoritative cache root directory according to
// $CATL_CACHE_DIR, $XDG_CACHE_HOME, the user's home directory or a temporary directory.
func DefaultCacheDir() string {
	if dir := os.Getenv("CATL_CACHE_DIR"); dir != "" {
		return dir
	}
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "catl")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache", "catl")
	}
	return filepath.Join(os.TempDir(), "catl-cache")
}

// Resolver locates external catalog dependencies on disk without network access.
type Resolver struct {
	catalogRoot string
	cacheDir    string
	vendorDir   string
}

// Option configures a Resolver instance.
type Option func(*Resolver)

// WithCacheDir configures a custom cache directory for the resolver.
func WithCacheDir(dir string) Option {
	return func(r *Resolver) {
		if dir != "" {
			r.cacheDir = dir
		}
	}
}

// WithVendorDir configures a custom vendor directory for the resolver.
func WithVendorDir(dir string) Option {
	return func(r *Resolver) {
		if dir != "" {
			r.vendorDir = dir
		}
	}
}

// NewResolver constructs a Resolver for the catalog at catalogRoot.
func NewResolver(catalogRoot string, opts ...Option) *Resolver {
	r := &Resolver{
		catalogRoot: catalogRoot,
		cacheDir:    DefaultCacheDir(),
		vendorDir:   filepath.Join(catalogRoot, ".catl", "vendor"),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// CatalogRoot returns the root directory of the owning catalog.
func (r *Resolver) CatalogRoot() string {
	return r.catalogRoot
}

// CacheDir returns the active cache directory.
func (r *Resolver) CacheDir() string {
	return r.cacheDir
}

// VendorDir returns the active vendor directory.
func (r *Resolver) VendorDir() string {
	return r.vendorDir
}

// HasVendor reports whether a vendor directory is present on disk.
func (r *Resolver) HasVendor() bool {
	st, err := os.Stat(r.vendorDir)
	return err == nil && st.IsDir()
}

// ResolveAll verifies lockfile freshness and resolves all declared dependencies.
// If any remote dependency is unpinned or stale, it returns lock.ErrLockStale.
// If any dependency files are missing or tampered with, it returns the respective error.
func (r *Resolver) ResolveAll(declared map[string]conf.Dependency, lockFile *lock.File) (map[string]*Resolved, error) {
	if err := lock.VerifyFreshness(declared, lockFile); err != nil {
		return nil, err
	}

	if len(declared) == 0 {
		return map[string]*Resolved{}, nil
	}

	res := make(map[string]*Resolved, len(declared))
	for name, dep := range declared {
		var lockEntry *lock.Entry
		if dep.Path == "" && lockFile != nil {
			req, err := lock.CanonicalRequested(dep)
			if err != nil {
				return nil, fmt.Errorf("canonical requested for %q: %w", name, err)
			}
			if entry, ok := lockFile.FindEntry(req); ok {
				lockEntry = &entry
			}
		}

		resolved, err := r.Resolve(name, dep, lockEntry)
		if err != nil {
			return nil, err
		}
		res[name] = resolved
	}

	return res, nil
}

// Resolve locates a single dependency on disk without performing network operations.
func (r *Resolver) Resolve(name string, dep conf.Dependency, lockEntry *lock.Entry) (*Resolved, error) {
	if err := conf.ValidateIdent(name); err != nil {
		return nil, fmt.Errorf("invalid dependency identifier %q: %w", name, err)
	}

	src, err := dep.Source()
	if err != nil {
		return nil, fmt.Errorf("dependency %q: %w", name, err)
	}

	switch s := src.(type) {
	case conf.PathSource:
		return r.resolvePath(name, s)
	case conf.GitSource:
		if lockEntry == nil {
			return nil, fmt.Errorf("%w: dependency %q", ErrLockRequired, name)
		}
		return r.resolveGit(name, s, lockEntry)
	default:
		return nil, fmt.Errorf("unsupported source type %T for dependency %q", src, name)
	}
}

// resolvePath resolves a local path dependency in-place relative to the catalog root.
func (r *Resolver) resolvePath(name string, s conf.PathSource) (*Resolved, error) {
	target := filepath.Clean(filepath.Join(r.catalogRoot, filepath.FromSlash(s.Dir)))
	absPath, err := filepath.Abs(target)
	if err != nil {
		return nil, fmt.Errorf("resolve path for %q: %w", name, err)
	}

	st, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: local path %q for dependency %q does not exist", ErrMissing, target, name)
		}
		return nil, fmt.Errorf("stat path for dependency %q: %w", name, err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%w: path %q for dependency %q is not a directory", ErrMissing, target, name)
	}

	if err := verifyCatalogTree(name, absPath, "", nil); err != nil {
		return nil, err
	}

	return &Resolved{
		Name:   name,
		Path:   absPath,
		FS:     os.DirFS(absPath),
		IsPath: true,
	}, nil
}

// resolveGit delegates Git dependency lookup to strict vendor mode or the global cache.
func (r *Resolver) resolveGit(name string, s conf.GitSource, entry *lock.Entry) (*Resolved, error) {
	if r.HasVendor() {
		return r.resolveVendored(name, entry)
	}
	return r.resolveCached(name, s, entry)
}

// resolveVendored resolves a Git dependency strictly from .catl/vendor/<name>.
func (r *Resolver) resolveVendored(name string, entry *lock.Entry) (*Resolved, error) {
	target := filepath.Join(r.vendorDir, name)
	absPath, err := filepath.Abs(target)
	if err != nil {
		return nil, fmt.Errorf("resolve vendor path for %q: %w", name, err)
	}

	st, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf(
				"%w: dependency %q not found in .catl/vendor (%s); run 'catl vendor'",
				ErrMissing,
				name,
				target,
			)
		}
		return nil, fmt.Errorf("stat vendor directory for dependency %q: %w", name, err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%w: vendored path %q is not a directory", ErrMissing, target)
	}

	if err := verifyCatalogTree(name, absPath, entry.Tree, ErrTamperedVendor); err != nil {
		return nil, err
	}

	return &Resolved{
		Name:       name,
		Path:       absPath,
		FS:         os.DirFS(absPath),
		IsVendored: true,
		TreeHash:   entry.Tree,
	}, nil
}

// resolveCached resolves a Git dependency from the global user cache directory.
func (r *Resolver) resolveCached(name string, s conf.GitSource, entry *lock.Entry) (*Resolved, error) {
	key := TreeKey(entry.Resolved, s.Subdir)
	target := filepath.Join(r.cacheDir, "git", "trees", key)
	absPath, err := filepath.Abs(target)
	if err != nil {
		return nil, fmt.Errorf("resolve cache path for %q: %w", name, err)
	}

	st, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			shortCommit := entry.Resolved
			if len(shortCommit) > 8 {
				shortCommit = shortCommit[:8]
			}
			return nil, fmt.Errorf(
				"%w: dependency %q (commit %s) not found in cache (%s); run 'catl update'",
				ErrMissing,
				name,
				shortCommit,
				target,
			)
		}
		return nil, fmt.Errorf("stat cached directory for dependency %q: %w", name, err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%w: cached path %q is not a directory", ErrMissing, target)
	}

	if err := verifyCatalogTree(name, absPath, entry.Tree, ErrCorruptCache); err != nil {
		return nil, err
	}

	return &Resolved{
		Name:     name,
		Path:     absPath,
		FS:       os.DirFS(absPath),
		TreeHash: entry.Tree,
	}, nil
}

// verifyCatalogTree checks that absPath contains a valid catl.toml with a matching
// catalog.name, and verifies content integrity against wantTreeHash if non-empty.
func verifyCatalogTree(name, absPath, wantTreeHash string, mismatchErr error) error {
	confPath := filepath.Join(absPath, "catl.toml")
	cfg, err := conf.LoadFile(confPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: target directory %q is missing catl.toml", ErrMissing, absPath)
		}
		return fmt.Errorf("load config for dependency %q at %q: %w", name, absPath, err)
	}

	if cfg.Catalog.Name != name {
		return fmt.Errorf(
			"%w: dependency key %q does not match target catalog name %q in %s",
			ErrTargetMismatch,
			name,
			cfg.Catalog.Name,
			confPath,
		)
	}

	if wantTreeHash == "" {
		return nil
	}

	actualHash, err := hash.Tree(os.DirFS(absPath))
	if err != nil {
		return fmt.Errorf("calculate tree hash for %q: %w", name, err)
	}

	if actualHash != wantTreeHash {
		return fmt.Errorf(
			"%w: dependency %q tree hash is %s, want %s",
			mismatchErr,
			name,
			actualHash,
			wantTreeHash,
		)
	}

	return nil
}
