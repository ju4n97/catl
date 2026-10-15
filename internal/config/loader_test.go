package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ju4n97/catl/internal/config"
)

// TestParse verifies parsing of valid and invalid TOML payloads into File structures.
// It tests strict unknown-field detection, BOM rejection, and proper variant deserialization.
func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		toml      string
		wantErr   bool
		targetErr error
		check     func(t *testing.T, f *config.File)
	}{
		{
			name: "valid catalog with flat inline dependencies",
			toml: `
[catalog]
name = "platform-rules"

[collections.rules]
schema = "schemas/rules.json"

[dependencies]
infra = { git = "https://github.com/acme/infra", tag = "v2.0.0" }
policies = { url = "https://cdn.example.com/p.tar.gz", digest = "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" }
core = { path = "../core" }
`,
			check: func(t *testing.T, f *config.File) {
				t.Helper()
				if f.Catalog.Name != "platform-rules" {
					t.Fatalf("catalog name = %q, want platform-rules", f.Catalog.Name)
				}
				if len(f.Dependencies) != 3 {
					t.Fatalf("got %d dependencies, want 3", len(f.Dependencies))
				}
				infra := f.Dependencies["infra"]
				if infra.Git != "https://github.com/acme/infra" || infra.Tag != "v2.0.0" {
					t.Errorf("unexpected infra dependency: %+v", infra)
				}
				policies := f.Dependencies["policies"]
				if policies.URL != "https://cdn.example.com/p.tar.gz" {
					t.Errorf("unexpected policies dependency: %+v", policies)
				}
				core := f.Dependencies["core"]
				if core.Path != "../core" {
					t.Errorf("unexpected core dependency: %+v", core)
				}
			},
		},
		{
			name:    "rejects utf8 byte order mark",
			toml:    "\xef\xbb\xbf[catalog]\nname = \"bom-cat\"\n[collections.a]\n",
			wantErr: true,
		},
		{
			name: "rejects unknown keys at root level",
			toml: `
[catalog]
name = "my-cat"
unknown_field = 123

[collections.rules]
`,
			wantErr:   true,
			targetErr: config.ErrRequiresNewer,
		},
		{
			name: "rejects unknown properties in dependency tables",
			toml: `
[catalog]
name = "my-cat"

[collections.rules]

[dependencies.infra]
infra = { git = "https://github.com/acme/infra", tag = "v1", extra = true }
`,
			wantErr:   true,
			targetErr: config.ErrRequiresNewer,
		},
		{
			name: "rejects unexpected trailing document content",
			toml: `
[catalog]
name = "my-cat"

[collections.rules]
---
extra = "document"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := config.Parse([]byte(tt.toml))
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %t", err, tt.wantErr)
			}

			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Errorf("Parse() error = %v, want %v", err, tt.targetErr)
			}

			if tt.check != nil && err == nil {
				tt.check(t, f)
			}
		})
	}
}

// TestLoadFile tests reading and parsing a catl.toml file directly from the filesystem.
func TestLoadFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "catl.toml")

	content := `
[catalog]
name = "on-disk"

[collections.records]
`
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	f, err := config.LoadFile(filePath)
	if err != nil {
		t.Fatalf("LoadFile(%q) error = %v", filePath, err)
	}

	if f.Catalog.Name != "on-disk" {
		t.Errorf("unexpected loaded catalog: %+v", f.Catalog)
	}

	// Missing file should return an error
	_, err = config.LoadFile(filepath.Join(tmpDir, "missing.toml"))
	if err == nil {
		t.Error("LoadFile() on missing path expected error, got nil")
	}
}

// TestFileMarshal verifies that valid configurations can be marshaled and reparsed losslessly.
func TestFileMarshal(t *testing.T) {
	t.Parallel()

	original := `
[catalog]
name = "roundtrip-cat"

[collections.rules]
dir = "rules"
schema = "schemas/rules.json"

[dependencies]
infra = { git = "https://github.com/acme/infra", tag = "v1.0.0" }
`

	f, err := config.Parse([]byte(original))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	data, err := f.Marshal()
	if err != nil {
		t.Fatalf("Marshal() failed: %v", err)
	}

	reparsed, err := config.Parse(data)
	if err != nil {
		t.Fatalf("Parse() of marshaled output failed: %v\nOutput:\n%s", err, string(data))
	}

	if reparsed.Catalog.Name != f.Catalog.Name {
		t.Errorf("roundtrip name = %q, want %q", reparsed.Catalog.Name, f.Catalog.Name)
	}
	if len(reparsed.Dependencies) != len(f.Dependencies) {
		t.Errorf(
			"roundtrip dependency count = %d, want %d",
			len(reparsed.Dependencies),
			len(f.Dependencies),
		)
	}
}
