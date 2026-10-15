package lock_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/lock"
)

const (
	sampleTreeHash = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
	sampleCommit1  = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
	sampleCommit2  = "4fe2c9b0123bf01c4a0b1e5f7a3c9d2e8b4f6071"
)

// TestParse verifies strict TOML decoding, schema validation, and error detection.
func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		toml      string
		wantErr   bool
		targetErr error
		check     func(t *testing.T, f *lock.File)
	}{
		{
			name: "valid lockfile with git pins",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v2.0.0"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"

[[dependency]]
requested = "git+https://github.com/acme/shared#branch=main"
resolved = "4fe2c9b0123bf01c4a0b1e5f7a3c9d2e8b4f6071"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`,
			wantErr: false,
			check: func(t *testing.T, f *lock.File) {
				t.Helper()
				if f.Version != 1 {
					t.Errorf("f.Version = %d, want 1", f.Version)
				}
				if len(f.Dependencies) != 2 {
					t.Fatalf("got %d dependencies, want 2", len(f.Dependencies))
				}
				first := f.Dependencies[0]
				if first.Requested != "git+https://github.com/acme/infra#tag=v2.0.0" {
					t.Errorf("unexpected requested: %s", first.Requested)
				}
				if first.Resolved != sampleCommit1 {
					t.Errorf("unexpected resolved: %s", first.Resolved)
				}
			},
		},
		{
			name:      "rejects utf8 byte order mark",
			toml:      "\xef\xbb\xbfversion = 1\n",
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects unknown fields strictly",
			toml: `
version = 1
unknown_property = true
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects missing version",
			toml: `
[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects unsupported version",
			toml: `
version = 2
`,
			wantErr:   true,
			targetErr: lock.ErrUnsupportedVersion,
		},
		{
			name: "rejects missing requested string",
			toml: `
version = 1

[[dependency]]
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects missing resolved commit",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects short git commit hash",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "d6e89a7"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects non-hex characters in git commit hash",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f6071z"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects malformed tree hash format",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "invalid-tree-hash"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects uppercase characters in tree hash",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "sha256:9F2C1A7B8E4F5A6B7C8D9E0F1A2B3C4D5E6F7A8B9C0D1E2F3A4B5C6D7E8F9A0B"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects non-hex characters in tree hash",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0bz"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
		{
			name: "rejects duplicate requested entries",
			toml: `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1"
resolved = "4fe2c9b0123bf01c4a0b1e5f7a3c9d2e8b4f6071a"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`,
			wantErr:   true,
			targetErr: lock.ErrInvalidLockfile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := lock.Parse([]byte(tt.toml))
			if (err != nil) != tt.wantErr {
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

// TestMarshal verifies lossless serialization and deterministic sorting by requested URI.
func TestMarshal(t *testing.T) {
	t.Parallel()

	// Insert entries out of alphabetical order
	f := &lock.File{
		Version: 1,
		Dependencies: []lock.Entry{
			{
				Requested: "git+https://github.com/acme/z-repo#tag=v1.0.0",
				Resolved:  sampleCommit1,
				Tree:      sampleTreeHash,
			},
			{
				Requested: "git+https://github.com/acme/a-repo#tag=v1.0.0",
				Resolved:  sampleCommit2,
				Tree:      sampleTreeHash,
			},
		},
	}

	data, err := lock.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	reparsed, err := lock.Parse(data)
	if err != nil {
		t.Fatalf("Parse() on marshaled output failed: %v", err)
	}

	// Verify entries were sorted in byte-wise order:
	wantOrder := []string{
		"git+https://github.com/acme/a-repo#tag=v1.0.0",
		"git+https://github.com/acme/z-repo#tag=v1.0.0",
	}

	for i, entry := range reparsed.Dependencies {
		if entry.Requested != wantOrder[i] {
			t.Errorf("entry[%d].Requested = %q, want %q", i, entry.Requested, wantOrder[i])
		}
	}
}

// TestLoad verifies reading and parsing from the OS filesystem.
func TestLoad(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "catl.lock")

	validTOML := `
version = 1

[[dependency]]
requested = "git+https://github.com/acme/infra#tag=v1.0.0"
resolved = "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718"
tree = "sha256:9f2c1a7b8e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b"
`
	if err := os.WriteFile(filePath, []byte(validTOML), 0o600); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	f, err := lock.Load(filePath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if f.Version != 1 || len(f.Dependencies) != 1 {
		t.Errorf("unexpected loaded lockfile: %+v", f)
	}

	// Loading missing file returns error
	_, err = lock.Load(filepath.Join(tmpDir, "missing.lock"))
	if err == nil {
		t.Fatal("expected error on missing file, got nil")
	}
}

// TestLoadFS verifies reading and parsing from an abstract fs.FS filesystem.
func TestLoadFS(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"catl.lock": &fstest.MapFile{
			Data: []byte("version = 1\n"),
		},
	}

	f, err := lock.LoadFS(fsys, "catl.lock")
	if err != nil {
		t.Fatalf("LoadFS() unexpected error: %v", err)
	}
	if f.Version != 1 {
		t.Errorf("f.Version = %d, want 1", f.Version)
	}

	_, err = lock.LoadFS(fsys, "missing.lock")
	if err == nil {
		t.Fatal("expected error on missing lockfile, got nil")
	}
}

// TestSave verifies atomic disk serialization using temporary files.
func TestSave(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "catl.lock")

	f := &lock.File{
		Version: 1,
		Dependencies: []lock.Entry{
			{
				Requested: "git+https://github.com/acme/infra#tag=v1.0.0",
				Resolved:  sampleCommit1,
				Tree:      sampleTreeHash,
			},
		},
	}

	if err := lock.Save(targetPath, f); err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}

	loaded, err := lock.Load(targetPath)
	if err != nil {
		t.Fatalf("Load() failed on saved file: %v", err)
	}

	if len(loaded.Dependencies) != 1 || loaded.Dependencies[0].Requested != f.Dependencies[0].Requested {
		t.Errorf("saved content mismatch: %+v", loaded)
	}
}

// TestCanonicalRequested verifies canonical requested string generation across source types.
func TestCanonicalRequested(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dep     conf.Dependency
		want    string
		wantErr bool
	}{
		{
			name: "git dependency with tag",
			dep: conf.Dependency{
				Git: "https://github.com/acme/infra",
				Tag: "v2.0.0",
			},
			want: "git+https://github.com/acme/infra#tag=v2.0.0",
		},
		{
			name: "git dependency with branch",
			dep: conf.Dependency{
				Git:    "https://github.com/acme/infra",
				Branch: "main",
			},
			want: "git+https://github.com/acme/infra#branch=main",
		},
		{
			name: "git dependency with branch and subdir",
			dep: conf.Dependency{
				Git:    "https://github.com/acme/infra",
				Branch: "main",
				Subdir: "catalogs/base",
			},
			want: "git+https://github.com/acme/infra#branch=main&subdir=catalogs/base",
		},
		{
			name: "git dependency with rev",
			dep: conf.Dependency{
				Git: "https://github.com/acme/infra",
				Rev: "d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718",
			},
			want: "git+https://github.com/acme/infra#rev=d6e89a7123bf01c4a0b1e5f7a3c9d2e8b4f60718",
		},
		{
			name: "local path dependency returns empty string",
			dep: conf.Dependency{
				Path: "../core",
			},
			want: "",
		},
		{
			name:    "empty dependency returns error",
			dep:     conf.Dependency{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := lock.CanonicalRequested(tt.dep)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CanonicalRequested() error = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("CanonicalRequested() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFind verifies lookup by requested URI key.
func TestFind(t *testing.T) {
	t.Parallel()

	f := &lock.File{
		Version: 1,
		Dependencies: []lock.Entry{
			{
				Requested: "git+https://github.com/acme/infra#tag=v1.0.0",
				Resolved:  sampleCommit1,
				Tree:      sampleTreeHash,
			},
		},
	}

	t.Run("find existing entry", func(t *testing.T) {
		t.Parallel()

		ent, ok := f.Find("git+https://github.com/acme/infra#tag=v1.0.0")
		if !ok {
			t.Fatal("expected entry to be found")
		}
		if ent.Resolved != sampleCommit1 {
			t.Errorf("ent.Resolved = %q, want %q", ent.Resolved, sampleCommit1)
		}
	})

	t.Run("find non-existent entry returns false", func(t *testing.T) {
		t.Parallel()

		_, ok := f.Find("git+https://github.com/acme/missing#tag=v1.0.0")
		if ok {
			t.Error("expected non-existent entry to return false")
		}
	})
}

// TestVerifyFreshness verifies lockfile staleness checks against declared dependencies.
func TestVerifyFreshness(t *testing.T) {
	t.Parallel()

	gitDep1 := conf.Dependency{Git: "https://github.com/acme/infra", Tag: "v1.0.0"}
	gitDep2 := conf.Dependency{Git: "https://github.com/acme/shared", Branch: "main"}
	pathDep := conf.Dependency{Path: "../core"}

	gitReq1, _ := lock.CanonicalRequested(gitDep1)
	gitReq2, _ := lock.CanonicalRequested(gitDep2)

	freshLock := &lock.File{
		Version: 1,
		Dependencies: []lock.Entry{
			{Requested: gitReq1, Resolved: sampleCommit1, Tree: sampleTreeHash},
			{Requested: gitReq2, Resolved: sampleCommit2, Tree: sampleTreeHash},
		},
	}

	t.Run("fresh lockfile matches declared dependencies", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"infra":  gitDep1,
			"shared": gitDep2,
			"core":   pathDep, // path dependency is ignored
		}

		if err := lock.VerifyFreshness(declared, freshLock); err != nil {
			t.Fatalf("VerifyFreshness() unexpected error: %v", err)
		}
	})

	t.Run("missing lockfile when remote dependencies exist returns err lock stale", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"infra": gitDep1,
		}

		err := lock.VerifyFreshness(declared, nil)
		if err == nil {
			t.Fatal("expected error for missing lockfile, got nil")
		}
		if !errors.Is(err, lock.ErrLockStale) {
			t.Errorf("got error %v, want ErrLockStale", err)
		}
	})

	t.Run("missing dependency pin in lockfile returns err lock stale", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"infra":  gitDep1,
			"shared": gitDep2,
			"extra":  {Git: "https://github.com/acme/extra", Tag: "v1.0.0"},
		}

		err := lock.VerifyFreshness(declared, freshLock)
		if err == nil {
			t.Fatal("expected error for missing pin, got nil")
		}
		if !errors.Is(err, lock.ErrLockStale) {
			t.Errorf("got error %v, want ErrLockStale", err)
		}
	})

	t.Run("extraneous pin in lockfile returns err lock stale", func(t *testing.T) {
		t.Parallel()

		// Declares only infra, but lockfile still contains shared
		declared := map[string]conf.Dependency{
			"infra": gitDep1,
		}

		err := lock.VerifyFreshness(declared, freshLock)
		if err == nil {
			t.Fatal("expected error for extraneous pin, got nil")
		}
		if !errors.Is(err, lock.ErrLockStale) {
			t.Errorf("got error %v, want ErrLockStale", err)
		}
	})

	t.Run("pin count mismatch returns err lock stale", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"infra":  gitDep1,
			"shared": gitDep2,
		}

		lockWithExtraPin := &lock.File{
			Version: 1,
			Dependencies: []lock.Entry{
				{Requested: gitReq1, Resolved: sampleCommit1, Tree: sampleTreeHash},
				{Requested: gitReq2, Resolved: sampleCommit2, Tree: sampleTreeHash},
				{
					Requested: "git+https://github.com/acme/extra#tag=v1.0.0",
					Resolved:  sampleCommit1,
					Tree:      sampleTreeHash,
				},
			},
		}

		err := lock.VerifyFreshness(declared, lockWithExtraPin)
		if !errors.Is(err, lock.ErrLockStale) {
			t.Fatalf("VerifyFreshness() error = %v, want ErrLockStale", err)
		}

		want := "catl.lock contains 3 pins but catl.toml declares 2 remote dependencies"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("VerifyFreshness() error = %q, want it to contain %q", err, want)
		}
	})

	t.Run("catalog with only path dependencies passes without lockfile", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"local-core": pathDep,
		}

		if err := lock.VerifyFreshness(declared, nil); err != nil {
			t.Fatalf("expected path-only catalog to pass without lockfile: %v", err)
		}
	})

	t.Run("catalog with only path dependencies fails if extraneous lockfile exists", func(t *testing.T) {
		t.Parallel()

		declared := map[string]conf.Dependency{
			"local-core": pathDep,
		}

		err := lock.VerifyFreshness(declared, freshLock)
		if err == nil {
			t.Fatal("expected error when lockfile contains pins but no remote dependencies declared, got nil")
		}
		if !errors.Is(err, lock.ErrLockStale) {
			t.Errorf("got error %v, want ErrLockStale", err)
		}
	})
}
