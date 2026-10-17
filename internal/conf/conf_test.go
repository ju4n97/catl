package conf_test

import (
	"testing"

	"github.com/ju4n97/catl/internal/conf"
)

// TestFile_ValidateCollections verifies relational checks that cannot be expressed in JSON Schema,
// such as duplicate effective directory paths across collections.
func TestFile_ValidateCollections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		file    conf.File
		wantErr bool
	}{
		{
			name: "distinct explicit directories",
			file: conf.File{
				Catalog: conf.Catalog{Name: "my-cat"},
				Collections: map[string]conf.Collection{
					"coll-a": {Dir: "dir-a"},
					"coll-b": {Dir: "dir-b"},
				},
			},
		},
		{
			name: "distinct default directories",
			file: conf.File{
				Catalog: conf.Catalog{Name: "my-cat"},
				Collections: map[string]conf.Collection{
					"coll-a": {},
					"coll-b": {},
				},
			},
		},
		{
			name: "collision between explicit directory and default directory",
			file: conf.File{
				Catalog: conf.Catalog{Name: "my-cat"},
				Collections: map[string]conf.Collection{
					"services": {},
					"other":    {Dir: "services"},
				},
			},
			wantErr: true,
		},
		{
			name: "schema path with directory traversal",
			file: conf.File{
				Catalog: conf.Catalog{Name: "my-cat"},
				Collections: map[string]conf.Collection{
					"services": {Schema: "../escapes/schema.json"},
				},
			},
			wantErr: true,
		},
		{
			name: "normalizes leading current directory in schema and directory",
			file: conf.File{
				Catalog: conf.Catalog{Name: "my-cat"},
				Collections: map[string]conf.Collection{
					"services": {Schema: "./schemas/rules.json", Dir: "./storage/services"},
				},
			},
		},
		{
			name: "overlapping directories",
			file: conf.File{
				Catalog: conf.Catalog{Name: "my-cat"},
				Collections: map[string]conf.Collection{
					"telemetry":          {Dir: "telemetry"},
					"telemetry-monitors": {Dir: "./telemetry/monitors"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.file.ValidateCollections()
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %t", err, tt.wantErr)
			}
			if !tt.wantErr && tt.name == "normalizes leading current directory in schema and directory" {
				coll := tt.file.Collections["services"]
				if coll.Schema != "schemas/rules.json" || coll.Dir != "storage/services" {
					t.Errorf("normalized collection = %+v, want schema and dir without leading ./", coll)
				}
			}
		})
	}
}

// TestDependencySource verifies that Source properly discriminates between flat dependency
// fields and returns the expected validated concrete variant.
func TestDependencySource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dep        conf.Dependency
		wantErr    bool
		wantSource conf.Source
	}{
		{
			name: "valid git source with tag",
			dep: conf.Dependency{
				Git: "https://github.com/acme/infra",
				Tag: "v1.0.0",
			},
			wantSource: conf.GitSource{Repo: "https://github.com/acme/infra", Tag: "v1.0.0"},
		},
		{
			name: "valid git source with branch and subdir",
			dep: conf.Dependency{
				Git:    "https://github.com/acme/infra",
				Branch: "main",
				Subdir: "catalogs/base",
			},
			wantSource: conf.GitSource{
				Repo:   "https://github.com/acme/infra",
				Branch: "main",
				Subdir: "catalogs/base",
			},
		},
		{
			name: "normalizes leading current directory in subdir",
			dep: conf.Dependency{
				Git:    "https://github.com/acme/infra",
				Branch: "main",
				Subdir: "./catalogs/base",
			},
			wantSource: conf.GitSource{
				Repo:   "https://github.com/acme/infra",
				Branch: "main",
				Subdir: "catalogs/base",
			},
		},
		{
			name: "valid git source with rev",
			dep: conf.Dependency{
				Git: "https://github.com/acme/infra",
				Rev: "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718",
			},
			wantSource: conf.GitSource{
				Repo: "https://github.com/acme/infra",
				Rev:  "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718",
			},
		},
		{
			name: "valid path source",
			dep: conf.Dependency{
				Path: "../core",
			},
			wantSource: conf.PathSource{Dir: "../core"},
		},
		{
			name:    "empty dependency has no source configured",
			dep:     conf.Dependency{},
			wantErr: true,
		},
		{
			name: "multiple source types configured simultaneously",
			dep: conf.Dependency{
				Git:  "https://github.com/acme/infra",
				Tag:  "v1.0.0",
				Path: "../core",
			},
			wantErr: true,
		},
		{
			name: "git missing reference target",
			dep: conf.Dependency{
				Git: "https://github.com/acme/infra",
			},
			wantErr: true,
		},
		{
			name: "git with ambiguous ref targets",
			dep: conf.Dependency{
				Git:    "https://github.com/acme/infra",
				Tag:    "v1.0.0",
				Branch: "main",
			},
			wantErr: true,
		},
		{
			name: "git with invalid subdir",
			dep: conf.Dependency{
				Git:    "https://github.com/acme/infra",
				Tag:    "v1.0.0",
				Subdir: "../escapes",
			},
			wantErr: true,
		},
		{
			name: "git with empty normalized subdir",
			dep: conf.Dependency{
				Git:    "https://github.com/acme/infra",
				Branch: "main",
				Subdir: "./",
			},
			wantErr: true,
		},
		{
			name: "git with unsupported repository scheme",
			dep: conf.Dependency{
				Git: "http://github.com/acme/infra",
				Tag: "v1.0.0",
			},
			wantErr: true,
		},
		{
			name: "git repository containing only scheme",
			dep: conf.Dependency{
				Git: "https://",
				Tag: "v1.0.0",
			},
			wantErr: true,
		},
		{
			name: "git with uppercase revision",
			dep: conf.Dependency{
				Git: "https://github.com/acme/infra",
				Rev: "D6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718",
			},
			wantErr: true,
		},
		{
			name: "path with absolute directory",
			dep: conf.Dependency{
				Path: "/absolute/path/not/allowed",
			},
			wantErr: true,
		},
		{
			name: "path containing backslashes",
			dep: conf.Dependency{
				Path: "..\\core",
			},
			wantErr: true,
		},
		{
			name: "path source with a git-only field",
			dep: conf.Dependency{
				Path: "../core",
				Tag:  "v1.0.0",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src, err := tt.dep.Source()
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("Source() error = %v, wantErr %t", err, tt.wantErr)
			}

			if !tt.wantErr && src != tt.wantSource {
				t.Errorf("Source() = %+v, want %+v", src, tt.wantSource)
			}
		})
	}
}
