package scan_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/errdef"
	"github.com/ju4n97/catl/internal/scan"
)

// TestTreeScan verifies catalog tree scanning, including valid entity discovery,
// clean-checkout root enforcement, and bundle invariant validations.
func TestTreeScan(t *testing.T) {
	t.Parallel()

	t.Run("valid catalog", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml": &fstest.MapFile{
				Data: []byte("[catalog]\nname = \"my-cat\"\n[collections.services]\n"),
			},
			"catl.lock":      &fstest.MapFile{Data: []byte("version = 1\n")},
			".gitattributes": &fstest.MapFile{Data: []byte("* text=auto\n")},

			"services/auth.yaml":                 &fstest.MapFile{Data: []byte("enabled: true\n")},
			"services/billing/manifest.yaml":     &fstest.MapFile{Data: []byte("enabled: true\n")},
			"services/billing/runbook.md":        &fstest.MapFile{Data: []byte("# Runbook\n")},
			"services/billing/steps/step1.md":    &fstest.MapFile{Data: []byte("Step 1\n")},
			"services/billing/.hidden-draft.txt": &fstest.MapFile{Data: []byte("ignored\n")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "my-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
				"drafts":   {},
			},
		}

		res, err := scan.TreeScan(fsys, cfg)
		if err != nil {
			t.Fatalf("scan.TreeScan() unexpected error: %v", err)
		}

		services := res.Collections["services"]
		if len(services) != 2 {
			t.Fatalf("services count = %d, want 2", len(services))
		}

		auth := services[0]
		if auth.ID != "auth" || auth.Shape != scan.EntityShapeFlat ||
			auth.ManifestPath != "services/auth.yaml" {
			t.Errorf("unexpected auth entity: %+v", auth)
		}

		billing := services[1]
		if billing.ID != "billing" || billing.Shape != scan.EntityShapeBundle ||
			billing.ManifestPath != "services/billing/manifest.yaml" {
			t.Errorf("unexpected billing entity: %+v", billing)
		}
		if len(billing.CompanionFiles) != 2 {
			t.Fatalf("companion files count = %d, want 2", len(billing.CompanionFiles))
		}
		if billing.CompanionFiles[0] != "runbook.md" ||
			billing.CompanionFiles[1] != "steps/step1.md" {
			t.Errorf("unexpected companion files: %v", billing.CompanionFiles)
		}

		drafts := res.Collections["drafts"]
		if len(drafts) != 0 {
			t.Errorf("drafts count = %d, want 0", len(drafts))
		}
	})

	t.Run("stray root file", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":       &fstest.MapFile{Data: []byte("")},
			"services/a.yaml": &fstest.MapFile{Data: []byte("")},
			"scratch.txt":     &fstest.MapFile{Data: []byte("stray")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "my-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on stray root file, got nil")
		}
		if !errors.Is(err, errdef.ErrStrayRootEntry) {
			t.Errorf("got %v, want ErrStrayRootEntry", err)
		}
	})

	t.Run("stray root directory", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":       &fstest.MapFile{Data: []byte("")},
			"services/a.yaml": &fstest.MapFile{Data: []byte("")},
			"tmp/scratch.txt": &fstest.MapFile{Data: []byte("stray")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "my-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on stray root directory, got nil")
		}
		if !errors.Is(err, errdef.ErrStrayRootEntry) {
			t.Errorf("got %v, want ErrStrayRootEntry", err)
		}
	})

	t.Run("conflicting entity shapes", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":                   &fstest.MapFile{Data: []byte("")},
			"services/auth.yaml":          &fstest.MapFile{Data: []byte("")},
			"services/auth/manifest.yaml": &fstest.MapFile{Data: []byte("")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "test-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on conflicting entity shapes, got nil")
		}
		if !errors.Is(err, errdef.ErrConflictingEntity) {
			t.Errorf("got %v, want ErrConflictingEntity", err)
		}
	})

	t.Run("missing manifest", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":            &fstest.MapFile{Data: []byte("")},
			"services/auth/doc.md": &fstest.MapFile{Data: []byte("")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "test-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on missing bundle manifest, got nil")
		}
		if !errors.Is(err, errdef.ErrMissingManifest) {
			t.Errorf("got %v, want ErrMissingManifest", err)
		}
	})

	t.Run("ambiguous manifest", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":                   &fstest.MapFile{Data: []byte("")},
			"services/auth/manifest.yaml": &fstest.MapFile{Data: []byte("")},
			"services/auth/manifest.json": &fstest.MapFile{Data: []byte("")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "test-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on ambiguous manifest formats, got nil")
		}
		if !errors.Is(err, errdef.ErrAmbiguousManifest) {
			t.Errorf("got %v, want ErrAmbiguousManifest", err)
		}
	})

	t.Run("invalid manifest extension", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":          &fstest.MapFile{Data: []byte("")},
			"services/notes.txt": &fstest.MapFile{Data: []byte("")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "test-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on invalid manifest extension, got nil")
		}
		if !errors.Is(err, errdef.ErrInvalidManifestType) {
			t.Errorf("got %v, want ErrInvalidManifestType", err)
		}
	})

	t.Run("rejects symlink at root", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":       &fstest.MapFile{Data: []byte("")},
			"services/a.yaml": &fstest.MapFile{Data: []byte("")},
			"link":            &fstest.MapFile{Mode: fs.ModeSymlink | 0o777, Data: []byte("services/a.yaml")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "test-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on symlink at root, got nil")
		}
		if !errors.Is(err, errdef.ErrForbiddenSymlink) {
			t.Errorf("got %v, want ErrForbiddenSymlink", err)
		}
	})

	t.Run("rejects symlinked collection entry", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":          &fstest.MapFile{Data: []byte("")},
			"services/link":      &fstest.MapFile{Mode: fs.ModeSymlink | 0o777, Data: []byte("auth")},
			"services/auth.yaml": &fstest.MapFile{Data: []byte("")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "test-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on symlinked collection entry, got nil")
		}
		if !errors.Is(err, errdef.ErrForbiddenSymlink) {
			t.Errorf("got %v, want ErrForbiddenSymlink", err)
		}
	})

	t.Run("rejects symlinked file inside bundle", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.toml":                    &fstest.MapFile{Data: []byte("")},
			"services/auth/manifest.yaml":  &fstest.MapFile{Data: []byte("")},
			"services/auth/linked-runbook": &fstest.MapFile{Mode: fs.ModeSymlink | 0o777, Data: []byte("../billing/runbook.md")},
		}

		cfg := &conf.File{
			Catalog: conf.Catalog{Name: "test-cat"},
			Collections: map[string]conf.Collection{
				"services": {},
			},
		}

		_, err := scan.TreeScan(fsys, cfg)
		if err == nil {
			t.Fatal("expected error on symlinked file inside bundle, got nil")
		}
		if !errors.Is(err, errdef.ErrForbiddenSymlink) {
			t.Errorf("got %v, want ErrForbiddenSymlink", err)
		}
	})
}
