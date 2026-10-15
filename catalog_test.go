package catl_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl"
	"github.com/ju4n97/catl/internal/scan"
	"github.com/ju4n97/catl/internal/schema"
)

// TestOpenFS verifies in-memory catalog loading, manifest validation,
// qualified identifier resolution, and error reporting across subtests.
func TestOpenFS(t *testing.T) {
	t.Parallel()

	validFS := fstest.MapFS{
		"catl.toml": &fstest.MapFile{
			Data: []byte(
				"[catalog]\nname = \"app-flags\"\n\n[collections.flags]\nschema = \"schemas/flags.json\"\n",
			),
		},
		"schemas/flags.json": &fstest.MapFile{
			Data: []byte(`{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "object",
				"required": ["enabled"],
				"properties": {
					"enabled": { "type": "boolean" },
					"rollout": { "type": "integer", "maximum": 100 }
				},
				"additionalProperties": false
			}`),
		},
		"flags/dark-mode.yaml": &fstest.MapFile{
			Data: []byte("enabled: true\nrollout: 100\n"),
		},
		"flags/new-checkout/manifest.yaml": &fstest.MapFile{
			Data: []byte("enabled: false\nrollout: 10\n"),
		},
		"flags/new-checkout/notes.md": &fstest.MapFile{
			Data: []byte("# Experiment Notes\n"),
		},
	}

	ctx := context.Background()

	t.Run("open valid catalog", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validFS, catl.Options{})
		if err != nil {
			t.Fatalf("OpenFS() unexpected error: %v", err)
		}

		if cat.Name() != "app-flags" {
			t.Errorf("cat.Name() = %q, want 'app-flags'", cat.Name())
		}
		if len(cat.Entities()) != 2 {
			t.Errorf("expected 2 entities, got %d", len(cat.Entities()))
		}
	})

	t.Run("resolve entity by qualified id", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validFS, catl.Options{})
		if err != nil {
			t.Fatalf("OpenFS() unexpected error: %v", err)
		}

		ent, err := cat.Resolve("flags", "app-flags.dark-mode")
		if err != nil {
			t.Fatalf("Resolve() failed: %v", err)
		}

		if ent.QualifiedID() != "app-flags.dark-mode" {
			t.Errorf("QualifiedID() = %q, want 'app-flags.dark-mode'", ent.QualifiedID())
		}
		if ent.Shape() != scan.EntityShapeFlat {
			t.Errorf("Shape() = %v, want flat", ent.Shape())
		}
	})

	t.Run("resolve rejects bare identifier", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validFS, catl.Options{})
		if err != nil {
			t.Fatalf("OpenFS() unexpected error: %v", err)
		}

		_, err = cat.Resolve("flags", "dark-mode")
		if err == nil {
			t.Fatal("expected error on bare identifier, got nil")
		}
		if !errors.Is(err, catl.ErrNotQualified) {
			t.Errorf("got %v, want ErrNotQualified", err)
		}
	})

	t.Run("resolve returns not found for unknown entity or namespace", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validFS, catl.Options{})
		if err != nil {
			t.Fatalf("OpenFS() unexpected error: %v", err)
		}

		_, err = cat.Resolve("flags", "app-flags.unknown")
		if err == nil || !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("expected ErrNotFound for unknown entity, got %v", err)
		}

		_, err = cat.Resolve("flags", "other-ns.dark-mode")
		if err == nil || !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("expected ErrNotFound for unknown namespace, got %v", err)
		}
	})

	t.Run("collection entities filtering", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validFS, catl.Options{})
		if err != nil {
			t.Fatalf("OpenFS() unexpected error: %v", err)
		}

		ents, err := cat.CollectionEntities("flags")
		if err != nil {
			t.Fatalf("CollectionEntities('flags') failed: %v", err)
		}
		if len(ents) != 2 {
			t.Errorf("expected 2 entities in flags, got %d", len(ents))
		}

		_, err = cat.CollectionEntities("missing")
		if err == nil || !errors.Is(err, catl.ErrNotFound) {
			t.Errorf("expected ErrNotFound on missing collection, got %v", err)
		}
	})

	t.Run("schema validation failure bubbles up", func(t *testing.T) {
		t.Parallel()

		invalidFS := fstest.MapFS{
			"catl.toml": &fstest.MapFile{
				Data: []byte(
					"[catalog]\nname = \"app-flags\"\n\n[collections.flags]\nschema = \"schemas/flags.json\"\n",
				),
			},
			"schemas/flags.json": &fstest.MapFile{
				Data: []byte(`{"type": "object", "properties": {"rollout": {"maximum": 100}}}`),
			},
			"flags/bad.yaml": &fstest.MapFile{
				Data: []byte("rollout: 150\n"),
			},
		}

		_, err := catl.OpenFS(ctx, invalidFS, catl.Options{})
		if err == nil {
			t.Fatal("expected schema validation error, got nil")
		}

		var valErr *schema.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *schema.ValidationError, got %T", err)
		}
		if len(valErr.Diagnostics) != 1 {
			t.Fatalf("expected 1 diagnostic, got %d", len(valErr.Diagnostics))
		}
	})

	t.Run("catalog tree hash calculation", func(t *testing.T) {
		t.Parallel()

		cat, err := catl.OpenFS(ctx, validFS, catl.Options{})
		if err != nil {
			t.Fatalf("OpenFS() failed: %v", err)
		}

		h, err := cat.TreeHash()
		if err != nil {
			t.Fatalf("TreeHash() failed: %v", err)
		}
		if !strings.HasPrefix(h, "sha256:") || len(h) != 71 {
			t.Errorf("unexpected tree hash format: %q", h)
		}
	})
}
