package hash_test

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl/internal/errdef"
	"github.com/ju4n97/catl/internal/hash"
)

// TestTree verifies deterministic catalog tree hash calculation across
// empty trees, exclusions, lexicographical ordering, and invalid characters.
func TestTree(t *testing.T) {
	t.Parallel()

	t.Run("empty catalog matches known zero digest", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{}
		digest, err := hash.Tree(fsys)
		if err != nil {
			t.Fatalf("Tree() unexpected error: %v", err)
		}
		if digest != hash.EmptyTreeHash {
			t.Errorf("got %q, want %q", digest, hash.EmptyTreeHash)
		}
	})

	t.Run("catalog with only dotfiles and lockfile produces empty tree hash", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"catl.lock":                    &fstest.MapFile{Data: []byte("version = 1\n")},
			".gitattributes":               &fstest.MapFile{Data: []byte("* text=auto\n")},
			".catl/cache/archive.tar.gz":   &fstest.MapFile{Data: []byte("binary data")},
			"schemas/.hidden-service.json": &fstest.MapFile{Data: []byte("{}")},
		}

		digest, err := hash.Tree(fsys)
		if err != nil {
			t.Fatalf("Tree() unexpected error: %v", err)
		}
		if digest != hash.EmptyTreeHash {
			t.Errorf("got %q, want %q", digest, hash.EmptyTreeHash)
		}
	})

	t.Run("deterministic sorting regardless of map insertion order", func(t *testing.T) {
		t.Parallel()

		// Construct files out of alphabetical order
		fsys := fstest.MapFS{
			"services/billing.yaml": &fstest.MapFile{Data: []byte("title: Billing\n")},
			"catl.toml":             &fstest.MapFile{Data: []byte("[catalog]\nname = \"demo\"\n")},
			"schemas/service.json":  &fstest.MapFile{Data: []byte("{\"type\": \"object\"}\n")},
			"services/auth.yaml":    &fstest.MapFile{Data: []byte("title: Auth\n")},
		}

		digest1, err := hash.Tree(fsys)
		if err != nil {
			t.Fatalf("Tree() failed: %v", err)
		}

		// Manually calculate expected hash
		// Sorted order:
		// 1. catl.toml
		// 2. schemas/service.json
		// 3. services/auth.yaml
		// 4. services/billing.yaml
		h := sha256.New()
		files := []struct {
			path    string
			content []byte
		}{
			{"catl.toml", []byte("[catalog]\nname = \"demo\"\n")},
			{"schemas/service.json", []byte("{\"type\": \"object\"}\n")},
			{"services/auth.yaml", []byte("title: Auth\n")},
			{"services/billing.yaml", []byte("title: Billing\n")},
		}

		for _, f := range files {
			fHash := sha256.Sum256(f.content)
			fmt.Fprintf(h, "%x  %s\n", fHash, f.path)
		}
		wantDigest := fmt.Sprintf("sha256:%x", h.Sum(nil))

		if digest1 != wantDigest {
			t.Errorf("Tree() = %q, want %q", digest1, wantDigest)
		}
	})

	t.Run("rejects path with backslash", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"schemas\\service.json": &fstest.MapFile{Data: []byte("{}")},
		}

		_, err := hash.Tree(fsys)
		if err == nil {
			t.Fatal("expected error on backslash path, got nil")
		}
		if !errors.Is(err, errdef.ErrInvalidPath) {
			t.Errorf("got error %v, want ErrInvalidPath", err)
		}
	})

	t.Run("rejects path with ascii control characters", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"services/bad\x1fname.yaml": &fstest.MapFile{Data: []byte("bad")},
		}

		_, err := hash.Tree(fsys)
		if err == nil {
			t.Fatal("expected error on control character path, got nil")
		}
		if !errors.Is(err, errdef.ErrInvalidPath) {
			t.Errorf("got error %v, want ErrInvalidPath", err)
		}
	})

	t.Run("rejects symlink entries", func(t *testing.T) {
		t.Parallel()

		fsys := fstest.MapFS{
			"schemas/redirect": &fstest.MapFile{Mode: fs.ModeSymlink | 0o777, Data: []byte("service.json")},
		}

		_, err := hash.Tree(fsys)
		if err == nil {
			t.Fatal("expected error on symlink entry, got nil")
		}
		if !errors.Is(err, errdef.ErrForbiddenSymlink) {
			t.Errorf("got error %v, want ErrForbiddenSymlink", err)
		}
	})
}
