package scan_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl/internal/scan"
)

// TestEntityFS verifies scoped filesystem extraction for bundle entities and rejection for flat entities.
func TestEntityFS(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"services/billing/manifest.yaml": &fstest.MapFile{Data: []byte("title: Billing")},
		"services/billing/runbook.md":    &fstest.MapFile{Data: []byte("# Runbook")},
	}

	t.Run("bundle entity", func(t *testing.T) {
		t.Parallel()

		bundle := &scan.Entity{
			ID:         "billing",
			Collection: "services",
			Shape:      scan.EntityShapeBundle,
			BundleDir:  "services/billing",
		}

		bundleFS, err := bundle.FS(fsys)
		if err != nil {
			t.Fatalf("bundle.FS() failed: %v", err)
		}

		data, err := fs.ReadFile(bundleFS, "runbook.md")
		if err != nil {
			t.Fatalf("ReadFile() failed: %v", err)
		}
		if string(data) != "# Runbook" {
			t.Errorf("got %q, want '# Runbook'", string(data))
		}
	})

	t.Run("flat entity", func(t *testing.T) {
		t.Parallel()

		flat := &scan.Entity{
			ID:         "auth",
			Collection: "services",
			Shape:      scan.EntityShapeFlat,
		}

		if _, err := flat.FS(fsys); err == nil {
			t.Error("flat.FS() expected error, got nil")
		}
	})
}
