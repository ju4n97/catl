package scan_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl/internal/scan"
)

// TestEntityShapeString verifies that EntityShape variants produce the expected
// string representations.
func TestEntityShapeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		shape scan.EntityShape
		want  string
	}{
		{
			name:  "flat shape",
			shape: scan.EntityShapeFlat,
			want:  "flat",
		},
		{
			name:  "bundle shape",
			shape: scan.EntityShapeBundle,
			want:  "bundle",
		},
		{
			name:  "unknown shape",
			shape: scan.EntityShape(99),
			want:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.shape.String(); got != tt.want {
				t.Errorf("EntityShape.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

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
