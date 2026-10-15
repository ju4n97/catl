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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.file.ValidateCollections()
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %t", err, tt.wantErr)
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
			name: "valid archive source",
			dep: conf.Dependency{
				URL:    "https://cdn.example.com/data.tar.gz",
				Digest: "sha256:123",
			},
			wantSource: conf.ArchiveSource{
				URL:    "https://cdn.example.com/data.tar.gz",
				Digest: "sha256:123",
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
			name: "archive with insecure http url",
			dep: conf.Dependency{
				URL: "http://insecure.example.com/archive.tar.gz",
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

// TestNewDependency verifies helper conversion from typed Source variants
// back into flat Dependency structs for serialization.
func TestNewDependency(t *testing.T) {
	t.Parallel()

	gitSrc := conf.GitSource{Repo: "https://github.com/acme/infra", Tag: "v1.0.0"}
	dep, err := conf.NewDependency(gitSrc)
	if err != nil {
		t.Fatalf("NewDependency(GitSource) error = %v", err)
	}
	if dep.Git != gitSrc.Repo || dep.Tag != gitSrc.Tag {
		t.Errorf("unexpected dependency fields: %+v", dep)
	}

	archSrc := conf.ArchiveSource{URL: "https://cdn.example.com/p.tar.gz", Digest: "sha256:abc"}
	dep, err = conf.NewDependency(archSrc)
	if err != nil {
		t.Fatalf("NewDependency(ArchiveSource) error = %v", err)
	}
	if dep.URL != archSrc.URL || dep.Digest != archSrc.Digest {
		t.Errorf("unexpected dependency fields: %+v", dep)
	}

	pathSrc := conf.PathSource{Dir: "../core"}
	dep, err = conf.NewDependency(pathSrc)
	if err != nil {
		t.Fatalf("NewDependency(PathSource) error = %v", err)
	}
	if dep.Path != pathSrc.Dir {
		t.Errorf("unexpected dependency fields: %+v", dep)
	}
}
