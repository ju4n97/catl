package config_test

import (
	"testing"

	"github.com/ju4n97/catl/internal/config"
)

// TestValidateSchema verifies that the embedded JSON Schema Draft 2020-12
// correctly enforces table exclusivity, identifier syntax, and nested dependency variants.
func TestValidateSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "valid catalog configuration",
			data: map[string]any{
				"catalog": map[string]any{
					"name": "valid-catalog",
				},
				"collections": map[string]any{
					"rules": map[string]any{
						"schema": "schemas/rules.json",
					},
				},
			},
		},
		{
			name: "valid workspace configuration",
			data: map[string]any{
				"workspace": map[string]any{
					"members": []any{"member-a", "member-b"},
				},
			},
		},
		{
			name: "rejects declaring both catalog and workspace",
			data: map[string]any{
				"catalog": map[string]any{
					"name": "catalog-name",
				},
				"workspace": map[string]any{
					"members": []any{"member-a"},
				},
			},
			wantErr: true,
		},
		{
			name: "rejects catalog name with uppercase letters",
			data: map[string]any{
				"catalog": map[string]any{
					"name": "InvalidName",
				},
				"collections": map[string]any{
					"rules": map[string]any{},
				},
			},
			wantErr: true,
		},
		{
			name: "rejects catalog without collections",
			data: map[string]any{
				"catalog": map[string]any{
					"name": "valid-name",
				},
				"collections": map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "valid flat git dependency with tag",
			data: map[string]any{
				"catalog":     map[string]any{"name": "test-cat"},
				"collections": map[string]any{"a": map[string]any{}},
				"dependencies": map[string]any{
					"infra": map[string]any{
						"git": "https://github.com/acme/infra",
						"tag": "v1.0.0",
					},
				},
			},
		},
		{
			name: "rejects git dependency with multiple refs",
			data: map[string]any{
				"catalog":     map[string]any{"name": "test-cat"},
				"collections": map[string]any{"a": map[string]any{}},
				"dependencies": map[string]any{
					"infra": map[string]any{
						"git":    "https://github.com/acme/infra",
						"tag":    "v1.0.0",
						"branch": "main",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "rejects dependency setting both git and path",
			data: map[string]any{
				"catalog":     map[string]any{"name": "test-cat"},
				"collections": map[string]any{"a": map[string]any{}},
				"dependencies": map[string]any{
					"ambiguous": map[string]any{
						"git":  "https://github.com/acme/infra",
						"tag":  "v1.0.0",
						"path": "../local",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateSchema(tt.data)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("ValidateSchema() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

// TestFileValidate verifies relational checks that cannot be expressed in JSON Schema,
// such as duplicate effective directory paths across collections.
func TestFileValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		file    config.File
		wantErr bool
	}{
		{
			name: "distinct explicit directories",
			file: config.File{
				Catalog: &config.Catalog{Name: "my-cat"},
				Collections: map[string]config.Collection{
					"coll-a": {Dir: "dir-a"},
					"coll-b": {Dir: "dir-b"},
				},
			},
		},
		{
			name: "distinct default directories",
			file: config.File{
				Catalog: &config.Catalog{Name: "my-cat"},
				Collections: map[string]config.Collection{
					"coll-a": {},
					"coll-b": {},
				},
			},
		},
		{
			name: "collision between explicit directory and default directory",
			file: config.File{
				Catalog: &config.Catalog{Name: "my-cat"},
				Collections: map[string]config.Collection{
					"services": {},
					"other":    {Dir: "services"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.file.Validate()
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}
