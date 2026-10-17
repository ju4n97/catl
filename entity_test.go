package catl_test

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl"
)

type FlagConfig struct {
	Enabled bool `json:"enabled"`
	Rollout int  `json:"rollout"`
}

// TestEntity verifies entity metadata fields, unmarshaling behavior,
// and bundle companion filesystem access.
func TestEntity(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"catl.toml": &fstest.MapFile{
			Data: []byte("[catalog]\nname = \"test-cat\"\n\n[collections.flags]\n"),
		},
		"flags/simple.yaml": &fstest.MapFile{
			Data: []byte("enabled: true\nrollout: 50\n"),
		},
		"flags/complex/manifest.yaml": &fstest.MapFile{
			Data: []byte("enabled: false\nrollout: 0\n"),
		},
		"flags/complex/notes.md": &fstest.MapFile{
			Data: []byte("# Rollout Notes\n"),
		},
	}

	ctx := context.Background()
	cat, err := catl.OpenFS(ctx, fsys)
	if err != nil {
		t.Fatalf("OpenFS() failed: %v", err)
	}

	t.Run("flat entity direct fields and unmarshal", func(t *testing.T) {
		t.Parallel()

		ent, err := cat.Resolve("flags", "test-cat.simple")
		if err != nil {
			t.Fatalf("Resolve() failed: %v", err)
		}

		if ent.ID != "simple" {
			t.Errorf("ID = %q, want 'simple'", ent.ID)
		}
		if ent.Collection != "flags" {
			t.Errorf("Collection = %q, want 'flags'", ent.Collection)
		}
		if ent.Namespace != "test-cat" {
			t.Errorf("Namespace = %q, want 'test-cat'", ent.Namespace)
		}
		if ent.QualifiedID != "test-cat.simple" {
			t.Errorf("QualifiedID = %q, want 'test-cat.simple'", ent.QualifiedID)
		}
		if ent.ManifestPath != "flags/simple.yaml" {
			t.Errorf("ManifestPath = %q, want 'flags/simple.yaml'", ent.ManifestPath)
		}
		if ent.IsBundle() {
			t.Errorf("IsBundle() = true, want false")
		}
		if len(ent.CompanionFiles) != 0 {
			t.Errorf("unexpected companion files: %v", ent.CompanionFiles)
		}

		var cfg FlagConfig
		if err := ent.Unmarshal(&cfg); err != nil {
			t.Fatalf("Unmarshal() failed: %v", err)
		}
		if !cfg.Enabled || cfg.Rollout != 50 {
			t.Errorf("unexpected unmarshaled config: %+v", cfg)
		}

		if _, err := ent.FS(); err == nil {
			t.Error("flat entity FS() expected error, got nil")
		}
	})

	t.Run("bundle entity companion filesystem", func(t *testing.T) {
		t.Parallel()

		ent, err := cat.Resolve("flags", "test-cat.complex")
		if err != nil {
			t.Fatalf("Resolve() failed: %v", err)
		}

		if !ent.IsBundle() {
			t.Errorf("IsBundle() = false, want true")
		}
		if len(ent.CompanionFiles) != 1 || ent.CompanionFiles[0] != "notes.md" {
			t.Errorf("unexpected companion files: %v", ent.CompanionFiles)
		}

		bundleFS, err := ent.FS()
		if err != nil {
			t.Fatalf("FS() failed: %v", err)
		}

		data, err := fs.ReadFile(bundleFS, "notes.md")
		if err != nil {
			t.Fatalf("ReadFile() from bundle failed: %v", err)
		}
		if string(data) != "# Rollout Notes\n" {
			t.Errorf("companion file content mismatch: %q", string(data))
		}
	})
}
