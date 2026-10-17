package schema_test

import (
	"bytes"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ju4n97/catl/internal/conf"
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

func compileSchema(url string, raw []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err := compiler.AddResource(url, doc); err != nil {
		return nil, err
	}

	return compiler.Compile(url)
}

func compileCatlSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	compiled, err := compileSchema(schema.CatlURL, schema.CatlJSON)
	if err != nil {
		t.Fatalf("compile embedded schema: %v", err)
	}
	return compiled
}

// TestCompileSchema verifies that the compiler helper rejects malformed JSON Schema bytes.
func TestCompileSchema(t *testing.T) {
	t.Parallel()

	if _, err := compileSchema("https://example.com/bad.json", []byte(`{ invalid `)); err == nil {
		t.Fatal("compileSchema() expected error on malformed JSON, got nil")
	}
}

// TestSchemaAndConfAgree verifies that catl.schema.json and conf.Parse accept and reject the same configurations.
func TestSchemaAndConfAgree(t *testing.T) {
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
core = { path = "../core" }
`,
		},
		{
			name: "valid catalog with leading current directory paths",
			toml: `
[catalog]
name = "platform-services"

[collections.services]
schema = "./schemas/services.json"
dir = "./services"

[dependencies.infra]
git = "https://github.com/acme/infra"
branch = "main"
subdir = "./catalogs/base"
`,
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
			name: "dependency with unsupported repository scheme",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.infra]
git = "http://github.com/acme/infra"
tag = "v1.0.0"
`,
			wantErr: true,
		},
		{
			name: "dependency with uppercase revision",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.infra]
git = "https://github.com/acme/infra"
rev = "D6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
`,
			wantErr: true,
		},
		{
			name: "invalid dependency identifier",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.invalid_name]
path = "../infra"
			`,
			wantErr: true,
		},
		{
			name: "path dependency with git-only field",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.core]
path = "../core"
tag = "v1"
`,
			wantErr: true,
		},
		{
			name: "path dependency with present empty git-only field",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.core]
path = "../core"
tag = ""
`,
			wantErr: true,
		},
		{
			name: "git dependency with empty tag and branch",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.infra]
git = "https://github.com/acme/infra"
tag = ""
branch = "main"
`,
			wantErr: true,
		},
		{
			name: "git repository containing only scheme",
			toml: `
[catalog]
name = "my-catalog"

[collections.services]

[dependencies.infra]
git = "https://"
tag = "v1"
`,
			wantErr: true,
		},
	}

	compiled := compileCatlSchema(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schemaErr := compiled.Validate(parseGeneric(t, tt.toml))
			_, confErr := conf.Parse([]byte(tt.toml))
			if (schemaErr != nil) != tt.wantErr {
				t.Errorf("JSON Schema validation error = %v, wantErr %t", schemaErr, tt.wantErr)
			}
			if (confErr != nil) != tt.wantErr {
				t.Errorf("conf.Parse() error = %v, wantErr %t", confErr, tt.wantErr)
			}
			if (schemaErr != nil) != (confErr != nil) {
				t.Errorf("JSON Schema and conf.Parse disagree: schema error = %v, conf error = %v", schemaErr, confErr)
			}
		})
	}
}
