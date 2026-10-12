package config_test

import (
	"testing"

	"github.com/ju4n97/catl/internal/config"
)

// TestDependencySource verifies pointer-union discrimination and validation
// across GitSource, ArchiveSource, and PathSource.
func TestDependencySource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dep        config.Dependency
		wantErr    bool
		wantSource config.Source
	}{
		{
			name: "valid git source with tag",
			dep: config.Dependency{
				Git: &config.GitSource{Repo: "https://github.com/acme/infra", Tag: "v1.0.0"},
			},
			wantSource: config.GitSource{Repo: "https://github.com/acme/infra", Tag: "v1.0.0"},
		},
		{
			name: "valid archive source",
			dep: config.Dependency{
				Archive: &config.ArchiveSource{URL: "https://cdn.example.com/data.tar.gz", Digest: "sha256:123"},
			},
			wantSource: config.ArchiveSource{URL: "https://cdn.example.com/data.tar.gz", Digest: "sha256:123"},
		},
		{
			name: "valid path source",
			dep: config.Dependency{
				Path: &config.PathSource{Dir: "../core"},
			},
			wantSource: config.PathSource{Dir: "../core"},
		},
		{
			name:    "empty dependency with no variant set",
			dep:     config.Dependency{},
			wantErr: true,
		},
		{
			name: "multiple variants set simultaneously",
			dep: config.Dependency{
				Git:  &config.GitSource{Repo: "https://github.com/acme/infra", Tag: "v1.0.0"},
				Path: &config.PathSource{Dir: "../core"},
			},
			wantErr: true,
		},
		{
			name: "git missing repo",
			dep: config.Dependency{
				Git: &config.GitSource{Tag: "v1.0.0"},
			},
			wantErr: true,
		},
		{
			name: "git with ambiguous ref targets",
			dep: config.Dependency{
				Git: &config.GitSource{Repo: "https://repo", Tag: "v1", Branch: "main"},
			},
			wantErr: true,
		},
		{
			name: "archive with non-https url",
			dep: config.Dependency{
				Archive: &config.ArchiveSource{URL: "http://insecure.example.com/data.tar.gz"},
			},
			wantErr: true,
		},
		{
			name: "path with absolute directory",
			dep: config.Dependency{
				Path: &config.PathSource{Dir: "/absolute/path"},
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

// TestNewDependency verifies helper constructor behavior for programmatic additions.
func TestNewDependency(t *testing.T) {
	t.Parallel()

	git := config.GitSource{Repo: "https://github.com/acme/infra", Tag: "v1.0.0"}
	dep, err := config.NewDependency(git)
	if err != nil {
		t.Fatalf("NewDependency(GitSource) error = %v", err)
	}
	if dep.Git == nil || *dep.Git != git {
		t.Errorf("expected wrapped GitSource, got %+v", dep)
	}

	path := config.PathSource{Dir: "../core"}
	dep, err = config.NewDependency(path)
	if err != nil {
		t.Fatalf("NewDependency(PathSource) error = %v", err)
	}
	if dep.Path == nil || *dep.Path != path {
		t.Errorf("expected wrapped PathSource, got %+v", dep)
	}
}

// TestDependencyClone verifies deep copying of dependency pointers.
func TestDependencyClone(t *testing.T) {
	t.Parallel()

	orig := config.Dependency{
		Git: &config.GitSource{Repo: "https://github.com/acme/infra", Tag: "v1.0.0"},
	}

	clone := orig.Clone()
	if clone.Git == nil || clone.Git == orig.Git {
		t.Fatal("clone did not create an independent pointer")
	}

	// Mutating original must not affect clone
	orig.Git.Tag = "v2.0.0"
	if clone.Git.Tag != "v1.0.0" {
		t.Errorf("clone mutated with original: got %q, want v1.0.0", clone.Git.Tag)
	}
}
