package catl_test

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ju4n97/catl"
)

// writeTestCatalog scaffolds a catalog on disk for registry testing.
func writeTestCatalog(t *testing.T, dir, name string) {
	t.Helper()

	catDir := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Join(catDir, "items"), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	toml := "[catalog]\nname = \"" + name + "\"\n[collections.items]\n"
	if err := os.WriteFile(filepath.Join(catDir, "catl.toml"), []byte(toml), 0o600); err != nil {
		t.Fatalf("write catl.toml failed: %v", err)
	}

	entity := "status: ok\n"
	if err := os.WriteFile(filepath.Join(catDir, "items", "entry.yaml"), []byte(entity), 0o600); err != nil {
		t.Fatalf("write entity failed: %v", err)
	}
}

// TestNewRegistry verifies in-memory registry creation, catalog retrieval,
// and deterministic ordering.
func TestNewRegistry(t *testing.T) {
	t.Parallel()

	cat1, err := catl.OpenFS(t.Context(), validCatalogFS(), catl.Options{})
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}

	reg := catl.NewRegistry(cat1)

	t.Run("names returns sorted slice", func(t *testing.T) {
		t.Parallel()

		names := reg.Names()
		if len(names) != 1 || names[0] != "platform" {
			t.Errorf("reg.Names() = %v, want ['platform']", names)
		}
	})

	t.Run("lookup existing catalog", func(t *testing.T) {
		t.Parallel()

		c, ok := reg.Catalog("platform")
		if !ok || c == nil {
			t.Fatal("expected catalog 'platform' to be found")
		}
		if c.Name() != "platform" {
			t.Errorf("c.Name() = %q, want 'platform'", c.Name())
		}
	})

	t.Run("lookup missing catalog returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := reg.Catalog("missing")
		if ok {
			t.Error("expected missing catalog to return false")
		}
	})

	t.Run("reload on in-memory registry returns error", func(t *testing.T) {
		t.Parallel()

		err := reg.Reload(t.Context())
		if err == nil {
			t.Fatal("expected error reloading in-memory registry, got nil")
		}
	})
}

// TestOpenRegistry verifies discovery of multiple catalog directories on disk,
// exclusion of dotfiles, and individual catalog retrieval.
func TestOpenRegistry(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()

	// Write two valid catalogs and ignored paths
	writeTestCatalog(t, rootDir, "clinic1")
	writeTestCatalog(t, rootDir, "clinic2")

	// Ignored non-catalog directory and dotfile
	if err := os.MkdirAll(filepath.Join(rootDir, ".hidden"), 0o755); err != nil {
		t.Fatalf("mkdir .hidden failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "README.md"), []byte("# Registry\n"), 0o600); err != nil {
		t.Fatalf("write README failed: %v", err)
	}

	reg, err := catl.OpenRegistry(t.Context(), rootDir, catl.Options{Offline: true})
	if err != nil {
		t.Fatalf("OpenRegistry() unexpected error: %v", err)
	}

	t.Run("discovers all valid catalogs in alphabetical order", func(t *testing.T) {
		t.Parallel()

		want := []string{"clinic1", "clinic2"}
		if !slices.Equal(reg.Names(), want) {
			t.Errorf("reg.Names() = %v, want %v", reg.Names(), want)
		}
	})

	t.Run("access catalog and resolve entity", func(t *testing.T) {
		t.Parallel()

		clinic1, ok := reg.Catalog("clinic1")
		if !ok {
			t.Fatal("failed to find catalog 'clinic1'")
		}

		ent, err := clinic1.Resolve("items", "clinic1.entry")
		if err != nil {
			t.Fatalf("clinic1.Resolve() failed: %v", err)
		}
		if ent.ID != "entry" {
			t.Errorf("ent.ID = %q, want 'entry'", ent.ID)
		}
	})

	t.Run("catalogs slice matches names ordering", func(t *testing.T) {
		t.Parallel()

		cats := reg.Catalogs()
		if len(cats) != 2 {
			t.Fatalf("got %d catalogs, want 2", len(cats))
		}
		if cats[0].Name() != "clinic1" || cats[1].Name() != "clinic2" {
			t.Errorf("unexpected catalogs order: %s, %s", cats[0].Name(), cats[1].Name())
		}
	})
}

// TestRegistryReload verifies that Reload updates catalogs in memory atomically
// and preserves the previous state if an update contains invalid files.
func TestRegistryReload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rootDir := t.TempDir()

	writeTestCatalog(t, rootDir, "clinic1")

	reg, err := catl.OpenRegistry(ctx, rootDir, catl.Options{Offline: true})
	if err != nil {
		t.Fatalf("OpenRegistry() failed: %v", err)
	}

	if len(reg.Names()) != 1 {
		t.Fatalf("initial count = %d, want 1", len(reg.Names()))
	}

	t.Run("reload dynamically discovers newly added catalog", func(t *testing.T) {
		writeTestCatalog(t, rootDir, "clinic2")

		if err := reg.Reload(ctx); err != nil {
			t.Fatalf("Reload() failed: %v", err)
		}

		want := []string{"clinic1", "clinic2"}
		if !slices.Equal(reg.Names(), want) {
			t.Errorf("after reload reg.Names() = %v, want %v", reg.Names(), want)
		}
	})

	t.Run("reload failure preserves existing catalogs in memory", func(t *testing.T) {
		// Introduce invalid configuration into a new catalog directory
		brokenDir := filepath.Join(rootDir, "broken")
		if err := os.MkdirAll(brokenDir, 0o755); err != nil {
			t.Fatalf("mkdir broken failed: %v", err)
		}
		if err := os.WriteFile(filepath.Join(brokenDir, "catl.toml"), []byte("[catalog\ninvalid"), 0o600); err != nil {
			t.Fatalf("write broken catl.toml failed: %v", err)
		}

		err := reg.Reload(ctx)
		if err == nil {
			t.Fatal("expected Reload() error on broken catalog, got nil")
		}

		// Verify that existing catalogs remain active and queryable
		want := []string{"clinic1", "clinic2"}
		if !slices.Equal(reg.Names(), want) {
			t.Errorf("after failed reload reg.Names() = %v, want %v", reg.Names(), want)
		}

		if _, ok := reg.Catalog("clinic1"); !ok {
			t.Error("catalog 'clinic1' was lost after failed reload")
		}
	})
}
