package schema_test

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/pelletier/go-toml/v2"

	"github.com/ju4n97/catl/schema"
)

// parseGeneric decodes raw TOML into generic map/slice primitives for schema validation.
func parseGeneric(t *testing.T, raw string) any {
	t.Helper()

	var generic any
	if err := toml.Unmarshal([]byte(raw), &generic); err != nil {
		t.Fatalf("failed to unmarshal test fixture: %v", err)
	}
	return generic
}

// TestValidateCatl verifies that catl.schema.json correctly enforces catalog configuration
// rules. Passing cases assert success; error cases snapshot the exact diagnostic output.
func TestValidateCatl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		toml    string
		wantErr bool
	}{
		{
			name: "valid minimal catalog",
			toml: `
[catalog]
name = "platform-services"

[collections.services]
`,
			wantErr: false,
		},
		{
			name: "valid catalog with dependencies and explicit paths",
			toml: `
[catalog]
name = "platform-services"

[collections.services]
schema = "schemas/services.json"
dir = "services"

[dependencies]
infra = { git = "https://github.com/acme/infra", tag = "v1.0.0" }
policies = { url = "https://cdn.example.com/p.tar.gz", digest = "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" }
core = { path = "../core" }
`,
			wantErr: false,
		},
		{
			name: "missing catalog table",
			toml: `
[collections.services]
`,
			wantErr: true,
		},
		{
			name: "missing collections table",
			toml: `
[catalog]
name = "empty-catalog"
`,
			wantErr: true,
		},
		{
			name: "catalog name with uppercase characters",
			toml: `
[catalog]
name = "InvalidName"

[collections.services]
`,
			wantErr: true,
		},
		{
			name: "catalog name starting with hyphen",
			toml: `
[catalog]
name = "-invalid-name"

[collections.services]
`,
			wantErr: true,
		},
		{
			name: "collection name with underscore",
			toml: `
[catalog]
name = "my-catalog"

[collections.invalid_collection]
`,
			wantErr: true,
		},
		{
			name: "unknown property at root level",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

workspace = "legacy-workspace"
`,
			wantErr: true,
		},
		{
			name: "git dependency with conflicting tag and branch",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.infra]
git = "https://github.com/acme/infra"
tag = "v1.0.0"
branch = "main"
`,
			wantErr: true,
		},
		{
			name: "dependency with conflicting source types",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.infra]
git = "https://github.com/acme/infra"
tag = "v1.0.0"
path = "../infra"
`,
			wantErr: true,
		},
		{
			name: "archive dependency with insecure http url",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.infra]
url = "http://insecure.example.com/archive.tar.gz"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := parseGeneric(t, tt.toml)
			err := schema.ValidateCatl(data)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCatl() error = %v, wantErr %t", err, tt.wantErr)
			}

			if tt.wantErr {
				snaps.MatchSnapshot(t, err.Error())
			}
		})
	}
}

// TestCompile verifies that the generic Compile helper returns an error on malformed JSON
// and succeeds on valid schema bytes.
func TestCompile(t *testing.T) {
	t.Parallel()

	t.Run("compile valid schema", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "string"}`)
		sch, err := schema.Compile("https://example.com/test.json", raw)
		if err != nil {
			t.Fatalf("Compile() unexpected error: %v", err)
		}
		if sch == nil {
			t.Fatal("Compile() returned nil schema")
		}
	})

	t.Run("compile invalid json syntax", func(t *testing.T) {
		t.Parallel()

		_, err := schema.Compile("https://example.com/bad.json", []byte(`{ invalid `))
		if err == nil {
			t.Fatal("Compile() expected error on malformed json, got nil")
		}
	})
}
