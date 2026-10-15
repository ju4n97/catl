package source_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/hash"
	"github.com/ju4n97/catl/internal/lock"
	"github.com/ju4n97/catl/internal/source"
)

// scaffoldCatalog writes a valid minimal catalog on disk and returns its calculated tree hash.
func scaffoldCatalog(t *testing.T, dir, name string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(dir, "services"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	catlTOML := "[catalog]\nname = \"" + name + "\"\n[collections.services]\n"
	if err := os.WriteFile(filepath.Join(dir, "catl.toml"), []byte(catlTOML), 0o600); err != nil {
		t.Fatalf("write catl.toml failed: %v", err)
	}

	manifest := "tier: critical\n"
	if err := os.WriteFile(filepath.Join(dir, "services", "auth.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write auth.yaml failed: %v", err)
	}

	treeHash, err := hash.Tree(os.DirFS(dir))
	if err != nil {
		t.Fatalf("hash.Tree failed: %v", err)
	}

	return treeHash
}

// TestTreeKey verifies deterministic cache key derivation across root and subdirectory checkouts.
func TestTreeKey(t *testing.T) {
	t.Parallel()

	commit := "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"

	tests := []struct {
		name   string
		subdir string
		want   string
	}{
		{
			name:   "empty subdir returns commit",
			subdir: "",
			want:   commit,
		},
		{
			name:   "dot subdir returns commit",
			subdir: ".",
			want:   commit,
		},
		{
			name:   "clean subdirectory produces commit with 16 hex hash",
			subdir: "catalogs/infra",
			want:   source.TreeKey(commit, "catalogs/infra"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := source.TreeKey(commit, tt.subdir)
			if tt.subdir == "" || tt.subdir == "." {
				if got != tt.want {
					t.Errorf("TreeKey() = %q, want %q", got, tt.want)
				}
			} else {
				if !strings.HasPrefix(got, commit+"-") || len(got) != 57 {
					t.Errorf("TreeKey() = %q, want length 57 with commit prefix", got)
				}
			}
		})
	}

	t.Run("normalizes leading and trailing slashes", func(t *testing.T) {
		t.Parallel()

		k1 := source.TreeKey(commit, "catalogs/infra")
		k2 := source.TreeKey(commit, "/catalogs/infra/")
		k3 := source.TreeKey(commit, "catalogs/infra/.")
		if k1 != k2 || k1 != k3 {
			t.Errorf("expected normalized keys to match: %s vs %s vs %s", k1, k2, k3)
		}
	})

	t.Run("distinct subdirectories produce distinct keys", func(t *testing.T) {
		t.Parallel()

		k1 := source.TreeKey(commit, "catalogs/infra")
		k2 := source.TreeKey(commit, "catalogs/core")
		if k1 == k2 {
			t.Errorf("expected keys to differ for different subdirs: %s == %s", k1, k2)
		}
	})
}

// TestResolverPathSource verifies local path dependency resolution and validations.
func TestResolverPathSource(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	catDir := filepath.Join(rootDir, "main-cat")
	depDir := filepath.Join(rootDir, "sibling-core")

	if err := os.MkdirAll(catDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	_ = scaffoldCatalog(t, depDir, "sibling-core")

	r := source.NewResolver(catDir)

	t.Run("resolves valid sibling catalog", func(t *testing.T) {
		t.Parallel()

		dep := conf.Dependency{Path: "../sibling-core"}
		res, err := r.Resolve("sibling-core", dep, nil)
		if err != nil {
			t.Fatalf("Resolve() unexpected error: %v", err)
		}

		if !res.IsPath || res.IsVendored {
			t.Errorf("unexpected flags: IsPath=%v, IsVendored=%v", res.IsPath, res.IsVendored)
		}
		if res.Name != "sibling-core" {
			t.Errorf("res.Name = %q, want sibling-core", res.Name)
		}
		if res.Path != depDir {
			t.Errorf("res.Path = %q, want %q", res.Path, depDir)
		}
		if res.TreeHash != "" {
			t.Errorf("expected empty TreeHash for path dependency, got %q", res.TreeHash)
		}
	})

	t.Run("missing directory returns err missing", func(t *testing.T) {
		t.Parallel()

		dep := conf.Dependency{Path: "../non-existent"}
		_, err := r.Resolve("sibling-core", dep, nil)
		if err == nil {
			t.Fatal("expected error on missing path, got nil")
		}
		if !errors.Is(err, source.ErrMissing) {
			t.Errorf("got %v, want ErrMissing", err)
		}
	})

	t.Run("target catalog name mismatch returns err target mismatch", func(t *testing.T) {
		t.Parallel()

		// Key is "core", but target catalog name is "sibling-core"
		dep := conf.Dependency{Path: "../sibling-core"}
		_, err := r.Resolve("core", dep, nil)
		if err == nil {
			t.Fatal("expected error on name mismatch, got nil")
		}
		if !errors.Is(err, source.ErrTargetMismatch) {
			t.Errorf("got %v, want ErrTargetMismatch", err)
		}
	})
}

// TestResolverVendoredMode verifies strict vendored resolution and tamper detection.
func TestResolverVendoredMode(t *testing.T) {
	t.Parallel()

	catDir := t.TempDir()
	vendorDir := filepath.Join(catDir, ".catl", "vendor", "infra")
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	validHash := scaffoldCatalog(t, vendorDir, "infra")
	commit := "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"

	r := source.NewResolver(catDir)
	if !r.HasVendor() {
		t.Fatal("expected HasVendor() = true")
	}

	t.Run("resolves valid vendored dependency", func(t *testing.T) {
		t.Parallel()

		dep := conf.Dependency{Git: "https://github.com/acme/infra", Tag: "v1.0.0"}
		entry := &lock.Entry{
			Requested: "git+https://github.com/acme/infra#tag=v1.0.0",
			Resolved:  commit,
			Tree:      validHash,
		}

		res, err := r.Resolve("infra", dep, entry)
		if err != nil {
			t.Fatalf("Resolve() unexpected error: %v", err)
		}

		if !res.IsVendored || res.IsPath {
			t.Errorf("unexpected flags: IsVendored=%v, IsPath=%v", res.IsVendored, res.IsPath)
		}
		if res.TreeHash != validHash {
			t.Errorf("res.TreeHash = %q, want %q", res.TreeHash, validHash)
		}
	})

	t.Run("missing vendored dependency returns err missing with vendor hint", func(t *testing.T) {
		t.Parallel()

		dep := conf.Dependency{Git: "https://github.com/acme/missing", Tag: "v1.0.0"}
		entry := &lock.Entry{
			Requested: "git+https://github.com/acme/missing#tag=v1.0.0",
			Resolved:  commit,
			Tree:      validHash,
		}

		_, err := r.Resolve("missing", dep, entry)
		if err == nil {
			t.Fatal("expected error on missing vendored dependency, got nil")
		}
		if !errors.Is(err, source.ErrMissing) {
			t.Errorf("got %v, want ErrMissing", err)
		}
		if !strings.Contains(err.Error(), "run 'catl vendor'") {
			t.Errorf("expected hint to run catl vendor, got: %v", err)
		}
	})

	t.Run("tampered vendored file returns err tampered vendor", func(t *testing.T) {
		t.Parallel()

		tamperedCatDir := t.TempDir()
		tamperedVendor := filepath.Join(tamperedCatDir, ".catl", "vendor", "infra")
		if err := os.MkdirAll(tamperedVendor, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		originalHash := scaffoldCatalog(t, tamperedVendor, "infra")

		// Modify a file in vendor to trigger hash mismatch
		authPath := filepath.Join(tamperedVendor, "services", "auth.yaml")
		if err := os.WriteFile(authPath, []byte("tier: tampered\n"), 0o600); err != nil {
			t.Fatalf("write tampered file failed: %v", err)
		}

		rTampered := source.NewResolver(tamperedCatDir)
		dep := conf.Dependency{Git: "https://github.com/acme/infra", Tag: "v1.0.0"}
		entry := &lock.Entry{
			Requested: "git+https://github.com/acme/infra#tag=v1.0.0",
			Resolved:  commit,
			Tree:      originalHash,
		}

		_, err := rTampered.Resolve("infra", dep, entry)
		if err == nil {
			t.Fatal("expected error on tampered vendor file, got nil")
		}
		if !errors.Is(err, source.ErrTamperedVendor) {
			t.Errorf("got %v, want ErrTamperedVendor", err)
		}
	})
}

// TestResolverCachedMode verifies global cache resolution, cache key routing, and corruption checks.
func TestResolverCachedMode(t *testing.T) {
	t.Parallel()

	catDir := t.TempDir() // No .catl/vendor directory present
	cacheDir := t.TempDir()

	commit := "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
	key := source.TreeKey(commit, "")
	cachedTreeDir := filepath.Join(cacheDir, "git", "trees", key)
	if err := os.MkdirAll(cachedTreeDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	validHash := scaffoldCatalog(t, cachedTreeDir, "infra")

	r := source.NewResolver(catDir, source.WithCacheDir(cacheDir))
	if r.HasVendor() {
		t.Fatal("expected HasVendor() = false")
	}

	t.Run("resolves cached dependency when tree hash matches", func(t *testing.T) {
		t.Parallel()

		dep := conf.Dependency{Git: "https://github.com/acme/infra", Tag: "v1.0.0"}
		entry := &lock.Entry{
			Requested: "git+https://github.com/acme/infra#tag=v1.0.0",
			Resolved:  commit,
			Tree:      validHash,
		}

		res, err := r.Resolve("infra", dep, entry)
		if err != nil {
			t.Fatalf("Resolve() unexpected error: %v", err)
		}

		if res.IsVendored || res.IsPath {
			t.Errorf("unexpected flags: IsVendored=%v, IsPath=%v", res.IsVendored, res.IsPath)
		}
		if res.Path != cachedTreeDir {
			t.Errorf("res.Path = %q, want %q", res.Path, cachedTreeDir)
		}
	})

	t.Run("resolves cached dependency with subdirectory", func(t *testing.T) {
		t.Parallel()

		subdir := "catalogs/infra"
		subKey := source.TreeKey(commit, subdir)
		subTreeDir := filepath.Join(cacheDir, "git", "trees", subKey)
		if err := os.MkdirAll(subTreeDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		subHash := scaffoldCatalog(t, subTreeDir, "infra-sub")

		dep := conf.Dependency{
			Git:    "https://github.com/acme/monorepo",
			Tag:    "v1.0.0",
			Subdir: subdir,
		}
		entry := &lock.Entry{
			Requested: "git+https://github.com/acme/monorepo#tag=v1.0.0&subdir=catalogs/infra",
			Resolved:  commit,
			Tree:      subHash,
		}

		res, err := r.Resolve("infra-sub", dep, entry)
		if err != nil {
			t.Fatalf("Resolve() unexpected error: %v", err)
		}
		if res.Path != subTreeDir {
			t.Errorf("res.Path = %q, want %q", res.Path, subTreeDir)
		}
	})

	t.Run("missing cache entry returns err missing with update hint", func(t *testing.T) {
		t.Parallel()

		dep := conf.Dependency{Git: "https://github.com/acme/infra", Tag: "v2.0.0"}
		entry := &lock.Entry{
			Requested: "git+https://github.com/acme/infra#tag=v2.0.0",
			Resolved:  "4fe2c9b0123bf01c4a0b1e5f7a3c9d2e8b4f6071",
			Tree:      validHash,
		}

		_, err := r.Resolve("infra", dep, entry)
		if err == nil {
			t.Fatal("expected error on missing cache entry, got nil")
		}
		if !errors.Is(err, source.ErrMissing) {
			t.Errorf("got %v, want ErrMissing", err)
		}
		if !strings.Contains(err.Error(), "run 'catl update'") {
			t.Errorf("expected hint to run catl update, got: %v", err)
		}
	})

	t.Run("corrupted cache entry returns err corrupt cache", func(t *testing.T) {
		t.Parallel()

		dep := conf.Dependency{Git: "https://github.com/acme/infra", Tag: "v1.0.0"}
		entry := &lock.Entry{
			Requested: "git+https://github.com/acme/infra#tag=v1.0.0",
			Resolved:  commit,
			Tree:      "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		}

		_, err := r.Resolve("infra", dep, entry)
		if err == nil {
			t.Fatal("expected error on corrupt cache entry, got nil")
		}
		if !errors.Is(err, source.ErrCorruptCache) {
			t.Errorf("got %v, want ErrCorruptCache", err)
		}
	})
}

// TestResolverResolveAll verifies freshness enforcement before resolving declared dependencies.
func TestResolverResolveAll(t *testing.T) {
	t.Parallel()

	catDir := t.TempDir()
	cacheDir := t.TempDir()

	commit := "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
	cachedTreeDir := filepath.Join(cacheDir, "git", "trees", source.TreeKey(commit, ""))
	if err := os.MkdirAll(cachedTreeDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	infraHash := scaffoldCatalog(t, cachedTreeDir, "infra")

	siblingDir := filepath.Join(catDir, "sibling-core")
	_ = scaffoldCatalog(t, siblingDir, "core")

	r := source.NewResolver(catDir, source.WithCacheDir(cacheDir))

	gitDep := conf.Dependency{Git: "https://github.com/acme/infra", Tag: "v1.0.0"}
	pathDep := conf.Dependency{Path: "./sibling-core"}
	gitReq, _ := lock.CanonicalRequested(gitDep)

	freshLock := &lock.File{
		Version: 1,
		Dependencies: []lock.Entry{
			{Requested: gitReq, Resolved: commit, Tree: infraHash},
		},
	}

	t.Run("resolves mixed local path and cached dependencies when fresh", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"infra": gitDep,
			"core":  pathDep,
		}

		res, err := r.ResolveAll(declared, freshLock)
		if err != nil {
			t.Fatalf("ResolveAll() unexpected error: %v", err)
		}

		if len(res) != 2 {
			t.Fatalf("got %d resolved dependencies, want 2", len(res))
		}
		if !res["infra"].IsVendored && res["infra"].IsPath {
			t.Errorf("unexpected infra flags: %+v", res["infra"])
		}
		if !res["core"].IsPath {
			t.Errorf("expected core to be path dependency: %+v", res["core"])
		}
	})

	t.Run("stale lockfile returns err lock stale without resolving", func(t *testing.T) {
		t.Parallel()

		// Declares extra dependency missing from lockfile
		declared := map[string]conf.Dependency{
			"infra": gitDep,
			"extra": {Git: "https://github.com/acme/extra", Tag: "v1.0.0"},
		}

		_, err := r.ResolveAll(declared, freshLock)
		if err == nil {
			t.Fatal("expected error for stale lockfile, got nil")
		}
		if !errors.Is(err, lock.ErrLockStale) {
			t.Errorf("got %v, want lock.ErrLockStale", err)
		}
	})

	t.Run("path only catalog succeeds without lockfile", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"core": pathDep,
		}

		res, err := r.ResolveAll(declared, nil)
		if err != nil {
			t.Fatalf("ResolveAll() unexpected error: %v", err)
		}
		if len(res) != 1 || res["core"] == nil {
			t.Errorf("unexpected resolved map: %+v", res)
		}
	})

	t.Run("empty declared dependencies returns empty map", func(t *testing.T) {
		t.Parallel()

		res, err := r.ResolveAll(nil, nil)
		if err != nil {
			t.Fatalf("ResolveAll(nil, nil) unexpected error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("got %d resolved, want 0", len(res))
		}
	})
}
