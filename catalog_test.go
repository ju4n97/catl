package catl_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl"
)

// validCatalogFS returns an in-memory catalog filesystem with multiple collections,
// explicit directory overrides, flat and bundle entities, companion files, and schemas.
func validCatalogFS() fstest.MapFS {
	return fstest.MapFS{
		"catl.toml": &fstest.MapFile{
			Data: []byte(`
[catalog]
name = "platform"

[collections.monitors]
schema = "schemas/monitor.json"

[collections.services]
schema = "schemas/service.json"
dir = "svcs"

[collections.drafts]
# Permissive collection without a schema
`),
		},
		"schemas/service.json": &fstest.MapFile{
			Data: []byte(`{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "object",
				"required": ["tier"],
				"properties": {
					"tier": { "type": "string" },
					"active": { "type": "boolean" },
					"threshold": { "type": "integer", "maximum": 100 }
				},
				"additionalProperties": false
			}`),
		},
		"schemas/monitor.json": &fstest.MapFile{
			Data: []byte(`{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "object",
				"required": ["interval_seconds"],
				"properties": {
					"interval_seconds": { "type": "integer", "minimum": 1 }
				},
				"additionalProperties": false
			}`),
		},
		"svcs/auth.yaml": &fstest.MapFile{
			Data: []byte("tier: core\nactive: true\nthreshold: 50\n"),
		},
		"svcs/billing/manifest.yaml": &fstest.MapFile{
			Data: []byte("tier: critical\nactive: true\n"),
		},
		"svcs/billing/runbook.md": &fstest.MapFile{
			Data: []byte("# Billing Runbook\n"),
		},
		"monitors/latency.yaml": &fstest.MapFile{
			Data: []byte("interval_seconds: 10\n"),
		},
		"monitors/uptime.yaml": &fstest.MapFile{
			Data: []byte("interval_seconds: 60\n"),
		},
	}
}

// TestOpenFS verifies in-memory catalog loading, schema validation,
// tree hash generation, and metadata discovery across subtests.
func TestOpenFS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("valid catalog loads all collections and entities", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validCatalogFS())
		if err != nil {
			t.Fatalf("OpenFS() unexpected error: %v", err)
		}

		if cat.Name() != "platform" {
			t.Errorf("cat.Name() = %q, want %q", cat.Name(), "platform")
		}
		if cat.RootPath() != "" {
			t.Errorf("cat.RootPath() = %q, want empty for OpenFS", cat.RootPath())
		}
		if cat.FS() == nil {
			t.Fatal("cat.FS() returned nil")
		}

		wantColls := []string{"drafts", "monitors", "services"}
		if !slices.Equal(cat.Collections(), wantColls) {
			t.Errorf("cat.Collections() = %v, want %v", cat.Collections(), wantColls)
		}

		all := cat.Entities()
		if len(all) != 4 {
			t.Fatalf("cat.Entities() count = %d, want 4", len(all))
		}
	})

	t.Run("tree hash returns valid sha256 digest", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validCatalogFS())
		if err != nil {
			t.Fatalf("OpenFS() failed: %v", err)
		}

		h, err := cat.TreeHash()
		if err != nil {
			t.Fatalf("TreeHash() unexpected error: %v", err)
		}

		if !strings.HasPrefix(h, "sha256:") || len(h) != 71 {
			t.Errorf("TreeHash() = %q, want valid sha256:... digest of length 71", h)
		}

		h2, err := cat.TreeHash()
		if err != nil {
			t.Fatalf("TreeHash() repeated call failed: %v", err)
		}
		if h != h2 {
			t.Errorf("TreeHash() non-deterministic: %q != %q", h, h2)
		}
	})

	t.Run("schema validation failure bubbles up", func(t *testing.T) {
		t.Parallel()

		fsys := validCatalogFS()
		fsys["svcs/auth.yaml"] = &fstest.MapFile{
			Data: []byte("tier: core\nactive: not-a-boolean\n"),
		}

		_, err := catl.OpenFS(ctx, fsys)
		if err == nil {
			t.Fatal("expected schema validation error, got nil")
		}

		var valErr *catl.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *catl.ValidationError, got %T: %v", err, err)
		}
		if len(valErr.Diagnostics) != 1 {
			t.Errorf("got %d diagnostics, want 1", len(valErr.Diagnostics))
		}
	})

	t.Run("malformed manifest returns decode error", func(t *testing.T) {
		t.Parallel()

		fsys := validCatalogFS()
		fsys["svcs/auth.yaml"] = &fstest.MapFile{
			Data: []byte("tier: [unclosed list\n"),
		}

		_, err := catl.OpenFS(ctx, fsys)
		if err == nil {
			t.Fatal("expected decode error on malformed YAML, got nil")
		}
	})

	t.Run("missing config file returns read error", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"services/item.yaml": &fstest.MapFile{Data: []byte("a: b\n")},
		}

		_, err := catl.OpenFS(ctx, fsys)
		if err == nil {
			t.Fatal("expected error on missing catl.toml, got nil")
		}
		if !strings.Contains(err.Error(), "catl.toml") {
			t.Errorf("expected error mentioning catl.toml, got %v", err)
		}
	})
}

// TestOpen verifies on-disk catalog loading, RootPath assignment, and error handling.
func TestOpen(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tmpDir := t.TempDir()

	configContent := "[catalog]\nname = \"local-cat\"\n[collections.items]\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "catl.toml"), []byte(configContent), 0o600); err != nil {
		t.Fatalf("write catl.toml failed: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(tmpDir, "items"), 0o755); err != nil {
		t.Fatalf("mkdir items failed: %v", err)
	}

	entityContent := "enabled: true\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "items", "first.yaml"), []byte(entityContent), 0o600); err != nil {
		t.Fatalf("write entity failed: %v", err)
	}

	t.Run("open on disk populates root path", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.Open(ctx, tmpDir)
		if err != nil {
			t.Fatalf("Open() unexpected error: %v", err)
		}

		if cat.RootPath() != tmpDir {
			t.Errorf("cat.RootPath() = %q, want %q", cat.RootPath(), tmpDir)
		}
		if cat.Name() != "local-cat" {
			t.Errorf("cat.Name() = %q, want %q", cat.Name(), "local-cat")
		}

		ents := cat.Entities()
		if len(ents) != 1 || ents[0].ID != "first" {
			t.Errorf("unexpected entities: %+v", ents)
		}
	})

	t.Run("open non-existent directory fails", func(t *testing.T) {
		t.Parallel()

		_, err := catl.Open(ctx, filepath.Join(tmpDir, "missing-dir"))
		if err == nil {
			t.Fatal("expected error opening missing directory, got nil")
		}
	})
}

// TestCatalogEntities verifies entity querying, collection-based filtering,
// and slice immutability.
func TestCatalogEntities(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cat, err := catl.OpenFS(ctx, validCatalogFS())
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}

	t.Run("all entities sorted by collection then id", func(t *testing.T) {
		t.Parallel()

		ents := cat.Entities()
		wantIDs := []string{"latency", "uptime", "auth", "billing"}
		wantColls := []string{"monitors", "monitors", "services", "services"}

		if len(ents) != len(wantIDs) {
			t.Fatalf("Entities() count = %d, want %d", len(ents), len(wantIDs))
		}

		for i := range len(ents) {
			if ents[i].ID != wantIDs[i] {
				t.Errorf("entity[%d].ID = %q, want %q", i, ents[i].ID, wantIDs[i])
			}
			if ents[i].Collection != wantColls[i] {
				t.Errorf("entity[%d].Collection = %q, want %q", i, ents[i].Collection, wantColls[i])
			}
		}
	})

	t.Run("filter entities by single collection", func(t *testing.T) {
		t.Parallel()

		monitors := cat.Entities("monitors")
		if len(monitors) != 2 {
			t.Fatalf("got %d monitors, want 2", len(monitors))
		}
		if monitors[0].ID != "latency" || monitors[1].ID != "uptime" {
			t.Errorf("unexpected monitors order: %s, %s", monitors[0].ID, monitors[1].ID)
		}

		svcs := cat.Entities("services")
		if len(svcs) != 2 {
			t.Fatalf("got %d services, want 2", len(svcs))
		}
		if svcs[0].ID != "auth" || svcs[1].ID != "billing" {
			t.Errorf("unexpected services order: %s, %s", svcs[0].ID, svcs[1].ID)
		}
	})

	t.Run("filter entities across multiple collections", func(t *testing.T) {
		t.Parallel()

		filtered := cat.Entities("services", "monitors")
		if len(filtered) != 4 {
			t.Fatalf("got %d entities, want 4", len(filtered))
		}
		if filtered[0].Collection != "services" || filtered[2].Collection != "monitors" {
			t.Errorf("unexpected collections ordering: %+v", filtered)
		}
	})

	t.Run("filter with nonexistent collection returns empty", func(t *testing.T) {
		t.Parallel()

		ents := cat.Entities("nonexistent")
		if len(ents) != 0 {
			t.Errorf("got %d entities for nonexistent collection, want 0", len(ents))
		}
	})

	t.Run("returned slices are cloned and immutable", func(t *testing.T) {
		t.Parallel()

		colls := cat.Collections()
		colls[0] = "mutated"
		if cat.Collections()[0] == "mutated" {
			t.Error("mutating Collections() slice affected catalog state")
		}

		ents := cat.Entities()
		ents[0] = nil
		if cat.Entities()[0] == nil {
			t.Error("mutating Entities() slice affected catalog state")
		}
	})
}

// TestCatalogResolve verifies qualified identifier resolution, bare identifier
// rejection, and error reporting on non-existent targets.
func TestCatalogResolve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cat, err := catl.OpenFS(ctx, validCatalogFS())
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}

	t.Run("resolve flat entity", func(t *testing.T) {
		t.Parallel()

		ent, err := cat.Resolve("services", "platform.auth")
		if err != nil {
			t.Fatalf("Resolve() unexpected error: %v", err)
		}

		if ent.ID != "auth" {
			t.Errorf("ID = %q, want 'auth'", ent.ID)
		}
		if ent.Collection != "services" {
			t.Errorf("Collection = %q, want 'services'", ent.Collection)
		}
		if ent.Namespace != "platform" {
			t.Errorf("Namespace = %q, want 'platform'", ent.Namespace)
		}
		if ent.QualifiedID != "platform.auth" {
			t.Errorf("QualifiedID = %q, want 'platform.auth'", ent.QualifiedID)
		}
		if ent.ManifestPath != "svcs/auth.yaml" {
			t.Errorf("ManifestPath = %q, want 'svcs/auth.yaml'", ent.ManifestPath)
		}
		if ent.IsBundle() {
			t.Error("IsBundle() = true, want false for flat entity")
		}

		type serviceManifest struct {
			Tier   string `json:"tier"`
			Active bool   `json:"active"`
		}
		var m serviceManifest
		if err := ent.Unmarshal(&m); err != nil {
			t.Fatalf("Unmarshal() failed: %v", err)
		}
		if m.Tier != "core" || !m.Active {
			t.Errorf("unexpected unmarshaled manifest: %+v", m)
		}
	})

	t.Run("resolve bundle entity", func(t *testing.T) {
		t.Parallel()

		ent, err := cat.Resolve("services", "platform.billing")
		if err != nil {
			t.Fatalf("Resolve() unexpected error: %v", err)
		}

		if ent.ID != "billing" {
			t.Errorf("ID = %q, want 'billing'", ent.ID)
		}
		if !ent.IsBundle() {
			t.Error("IsBundle() = false, want true for bundle entity")
		}
		if len(ent.CompanionFiles) != 1 || ent.CompanionFiles[0] != "runbook.md" {
			t.Errorf("unexpected companion files: %v", ent.CompanionFiles)
		}

		bundleFS, err := ent.FS(ctx)
		if err != nil {
			t.Fatalf("FS() failed: %v", err)
		}
		data, _ := os.ReadFile(filepath.Join(t.TempDir(), "dummy")) // compile check
		_ = data
		_ = bundleFS
	})

	t.Run("bare identifier fails with err not qualified", func(t *testing.T) {
		t.Parallel()

		_, err := cat.Resolve("services", "auth")
		if err == nil {
			t.Fatal("expected error on bare identifier, got nil")
		}
		if !errors.Is(err, catl.ErrNotQualified) {
			t.Errorf("got error %v, want ErrNotQualified", err)
		}
	})

	t.Run("unknown namespace fails with err not found", func(t *testing.T) {
		t.Parallel()

		_, err := cat.Resolve("services", "other-app.auth")
		if err == nil {
			t.Fatal("expected error on foreign namespace, got nil")
		}
		if !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("got error %v, want ErrNotFound", err)
		}
	})

	t.Run("unknown collection fails with err not found", func(t *testing.T) {
		t.Parallel()

		_, err := cat.Resolve("unknown-coll", "platform.auth")
		if err == nil {
			t.Fatal("expected error on missing collection, got nil")
		}
		if !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("got error %v, want ErrNotFound", err)
		}
	})

	t.Run("unknown entity id fails with err not found", func(t *testing.T) {
		t.Parallel()

		_, err := cat.Resolve("services", "platform.nonexistent")
		if err == nil {
			t.Fatal("expected error on missing entity, got nil")
		}
		if !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("got error %v, want ErrNotFound", err)
		}
	})
}

// TestCatalogDependencies verifies cross-namespace entity resolution, lockfile checking,
// and path prefixing across local path and vendored dependencies.
func TestCatalogDependencies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("resolves entities from local path dependency", func(t *testing.T) {
		t.Parallel()

		rootDir := t.TempDir()
		coreDir := filepath.Join(rootDir, "core")
		appDir := filepath.Join(rootDir, "app")

		// Scaffold dependency catalog 'core'
		if err := os.MkdirAll(filepath.Join(coreDir, "services"), 0o755); err != nil {
			t.Fatalf("mkdir core failed: %v", err)
		}
		coreTOML := "[catalog]\nname = \"core\"\n[collections.services]\n"
		if err := os.WriteFile(filepath.Join(coreDir, "catl.toml"), []byte(coreTOML), 0o600); err != nil {
			t.Fatalf("write core catl.toml failed: %v", err)
		}
		if err := os.WriteFile(filepath.Join(coreDir, "services", "postgres.yaml"), []byte("tier: primary\n"), 0o600); err != nil {
			t.Fatalf("write postgres.yaml failed: %v", err)
		}

		// Scaffold root catalog 'app' depending on '../core'
		if err := os.MkdirAll(filepath.Join(appDir, "services"), 0o755); err != nil {
			t.Fatalf("mkdir app failed: %v", err)
		}
		appTOML := "[catalog]\nname = \"app\"\n[collections.services]\n\n[dependencies]\ncore = { path = \"../core\" }\n"
		if err := os.WriteFile(filepath.Join(appDir, "catl.toml"), []byte(appTOML), 0o600); err != nil {
			t.Fatalf("write app catl.toml failed: %v", err)
		}
		if err := os.WriteFile(filepath.Join(appDir, "services", "api.yaml"), []byte("tier: web\n"), 0o600); err != nil {
			t.Fatalf("write api.yaml failed: %v", err)
		}

		cat, err := catl.Open(ctx, appDir)
		if err != nil {
			t.Fatalf("Open() unexpected error: %v", err)
		}

		if !slices.Equal(cat.Dependencies(), []string{"core"}) {
			t.Errorf("cat.Dependencies() = %v, want ['core']", cat.Dependencies())
		}

		// Resolve local entity
		localEnt, err := cat.Resolve("services", "app.api")
		if err != nil {
			t.Fatalf("Resolve(local) failed: %v", err)
		}
		if localEnt.ManifestPath != "services/api.yaml" {
			t.Errorf("localEnt.ManifestPath = %q, want 'services/api.yaml'", localEnt.ManifestPath)
		}

		// Resolve cross-namespace dependency entity
		depEnt, err := cat.Resolve("services", "core.postgres")
		if err != nil {
			t.Fatalf("Resolve(core.postgres) failed: %v", err)
		}
		if depEnt.QualifiedID != "core.postgres" {
			t.Errorf("depEnt.QualifiedID = %q, want 'core.postgres'", depEnt.QualifiedID)
		}
		if depEnt.Namespace != "core" {
			t.Errorf("depEnt.Namespace = %q, want 'core'", depEnt.Namespace)
		}
		if depEnt.ManifestPath != "../core/services/postgres.yaml" {
			t.Errorf("depEnt.ManifestPath = %q, want '../core/services/postgres.yaml'", depEnt.ManifestPath)
		}

		// Verify AllEntities includes both sorted by collection then qualified ID
		all := cat.AllEntities()
		if len(all) != 2 {
			t.Fatalf("AllEntities() count = %d, want 2", len(all))
		}
		if all[0].QualifiedID != "app.api" || all[1].QualifiedID != "core.postgres" {
			t.Errorf("unexpected AllEntities ordering: %s, %s", all[0].QualifiedID, all[1].QualifiedID)
		}
	})

	t.Run("resolves bundle companion file from dependency", func(t *testing.T) {
		t.Parallel()

		rootDir := t.TempDir()
		coreDir := filepath.Join(rootDir, "infra")
		appDir := filepath.Join(rootDir, "web")

		if err := os.MkdirAll(filepath.Join(coreDir, "monitors", "network"), 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		coreTOML := "[catalog]\nname = \"infra\"\n[collections.monitors]\n"
		_ = os.WriteFile(filepath.Join(coreDir, "catl.toml"), []byte(coreTOML), 0o600)
		_ = os.WriteFile(filepath.Join(coreDir, "monitors", "network", "manifest.yaml"), []byte("active: true\n"), 0o600)
		_ = os.WriteFile(filepath.Join(coreDir, "monitors", "network", "runbook.md"), []byte("# Network Runbook\n"), 0o600)

		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		appTOML := "[catalog]\nname = \"web\"\n[collections.pages]\n\n[dependencies]\ninfra = { path = \"../infra\" }\n"
		_ = os.WriteFile(filepath.Join(appDir, "catl.toml"), []byte(appTOML), 0o600)

		cat, err := catl.Open(ctx, appDir)
		if err != nil {
			t.Fatalf("Open() failed: %v", err)
		}

		ent, err := cat.Resolve("monitors", "infra.network")
		if err != nil {
			t.Fatalf("Resolve(infra.network) failed: %v", err)
		}
		if !ent.IsBundle() {
			t.Fatal("expected bundle entity")
		}

		bfs, err := ent.FS(ctx)
		if err != nil {
			t.Fatalf("ent.FS() failed: %v", err)
		}

		data, err := fs.ReadFile(bfs, "runbook.md")
		if err != nil {
			t.Fatalf("ReadFile(runbook.md) failed: %v", err)
		}
		if string(data) != "# Network Runbook\n" {
			t.Errorf("companion file content mismatch: %q", string(data))
		}
	})

	t.Run("undeclared namespace returns err not found", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validCatalogFS())
		if err != nil {
			t.Fatalf("OpenFS() failed: %v", err)
		}

		_, err = cat.Resolve("services", "undeclared.item")
		if err == nil {
			t.Fatal("expected error on undeclared namespace, got nil")
		}
		if !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

// TestCatalogSchema verifies retrieval of collection schemas and permissive mode handling.
func TestCatalogSchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cat, err := catl.OpenFS(ctx, validCatalogFS())
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}

	t.Run("existing collection returns raw schema bytes", func(t *testing.T) {
		t.Parallel()

		schemaBytes, err := cat.Schema("services")
		if err != nil {
			t.Fatalf("Schema('services') unexpected error: %v", err)
		}
		if !strings.Contains(string(schemaBytes), "monitor") && !strings.Contains(string(schemaBytes), "tier") {
			t.Errorf("unexpected schema content: %s", string(schemaBytes))
		}
	})

	t.Run("permissive collection without schema returns nil bytes with no error", func(t *testing.T) {
		t.Parallel()

		schemaBytes, err := cat.Schema("drafts")
		if err != nil {
			t.Fatalf("Schema('drafts') unexpected error: %v", err)
		}
		if schemaBytes != nil {
			t.Errorf("expected nil schema bytes for permissive collection, got %s", string(schemaBytes))
		}
	})

	t.Run("non-existent collection returns err not found", func(t *testing.T) {
		t.Parallel()

		_, err := cat.Schema("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent collection, got nil")
		}
		if !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

// TestCatalogValidateManifest verifies in-memory dry-run manifest linting.
func TestCatalogValidateManifest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cat, err := catl.OpenFS(ctx, validCatalogFS())
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}

	t.Run("valid manifest passes validation", func(t *testing.T) {
		t.Parallel()

		validYAML := "tier: standard\nactive: true\nthreshold: 80\n"
		if err := cat.ValidateManifest("services", []byte(validYAML)); err != nil {
			t.Fatalf("ValidateManifest() unexpected error: %v", err)
		}
	})

	t.Run("schema violation returns validation error with diagnostics", func(t *testing.T) {
		t.Parallel()

		invalidYAML := "tier: standard\nthreshold: 150\n" // threshold maximum is 100
		err := cat.ValidateManifest("services", []byte(invalidYAML))
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}

		var valErr *catl.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *catl.ValidationError, got %T: %v", err, err)
		}
		if len(valErr.Diagnostics) == 0 {
			t.Fatal("expected diagnostics in validation error, got 0")
		}
	})

	t.Run("malformed yaml syntax returns decode error", func(t *testing.T) {
		t.Parallel()

		malformed := "tier: [unclosed\n"
		err := cat.ValidateManifest("services", []byte(malformed))
		if err == nil {
			t.Fatal("expected syntax error on unclosed yaml mapping, got nil")
		}
	})

	t.Run("permissive collection passes arbitrary valid yaml", func(t *testing.T) {
		t.Parallel()

		arbitrary := "custom_field: 42\nnested:\n  ok: true\n"
		if err := cat.ValidateManifest("drafts", []byte(arbitrary)); err != nil {
			t.Fatalf("expected permissive collection to pass, got: %v", err)
		}
	})

	t.Run("non-existent collection returns err not found", func(t *testing.T) {
		t.Parallel()

		err := cat.ValidateManifest("nonexistent", []byte("a: b\n"))
		if err == nil {
			t.Fatal("expected error on non-existent collection, got nil")
		}
		if !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
