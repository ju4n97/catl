package manifest_test

import (
	"errors"
	"testing"

	"github.com/ju4n97/catl/internal/errdef"
	"github.com/ju4n97/catl/internal/manifest"
)

// TestDecodeManifest verifies YAML 1.2 core decoding rules, syntax limits,
// duplicate key rejection, and pointer line-mapping.
func TestDecodeManifest(t *testing.T) {
	t.Parallel()

	t.Run("strict yaml 1.2 boolean strings", func(t *testing.T) {
		t.Parallel()

		raw := "enabled: no\nthreshold: 50\n"
		m, err := manifest.Decode([]byte(raw), "rules/test.yaml")
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		dict, ok := m.Raw.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", m.Raw)
		}

		// In YAML 1.2, unquoted 'no' is parsed as a string, not a boolean
		val, ok := dict["enabled"].(string)
		if !ok || val != "no" {
			t.Errorf("enabled = %v (%T), want string 'no'", dict["enabled"], dict["enabled"])
		}
	})

	t.Run("find line by pointer", func(t *testing.T) {
		t.Parallel()

		raw := `title: Service Test
timeout: 300
settings:
  nested: true
endpoints:
  - primary.internal
  - secondary.internal
`
		m, err := manifest.Decode([]byte(raw), "test.yaml")
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		if line := m.FindLine("/title"); line != 1 {
			t.Errorf("line for /title = %d, want 1", line)
		}
		if line := m.FindLine("/timeout"); line != 2 {
			t.Errorf("line for /timeout = %d, want 2", line)
		}
		if line := m.FindLine("/settings"); line != 3 {
			t.Errorf("line for /settings = %d, want 3", line)
		}
		if line := m.FindLine("/settings/nested"); line != 4 {
			t.Errorf("line for /settings/nested = %d, want 4", line)
		}
		if line := m.FindLine("/endpoints"); line != 5 {
			t.Errorf("line for /endpoints = %d, want 5", line)
		}
		if line := m.FindLine("/endpoints/0"); line != 6 {
			t.Errorf("line for /endpoints/0 = %d, want 6", line)
		}
		if line := m.FindLine("/endpoints/1"); line != 7 {
			t.Errorf("line for /endpoints/1 = %d, want 7", line)
		}
	})

	t.Run("rejects crlf line endings", func(t *testing.T) {
		t.Parallel()

		raw := "title: Test\r\nenabled: true\r\n"
		_, err := manifest.Decode([]byte(raw), "test.yaml")
		if err == nil {
			t.Fatal("expected error on CRLF line endings, got nil")
		}
		if !errors.Is(err, errdef.ErrCRLFForbidden) {
			t.Errorf("got %v, want ErrCRLFForbidden", err)
		}
	})

	t.Run("rejects duplicate keys", func(t *testing.T) {
		t.Parallel()

		raw := "title: A\ntitle: B\n"
		_, err := manifest.Decode([]byte(raw), "dup.yaml")
		if err == nil {
			t.Fatal("expected error on duplicate keys, got nil")
		}
	})

	t.Run("rejects multi document yaml", func(t *testing.T) {
		t.Parallel()

		raw := "title: Doc1\n---\ntitle: Doc2\n"
		_, err := manifest.Decode([]byte(raw), "multi.yaml")
		if err == nil {
			t.Fatal("expected error on multi-document yaml, got nil")
		}
		if !errors.Is(err, errdef.ErrMultiDocument) {
			t.Errorf("got %v, want ErrMultiDocument", err)
		}
	})

	t.Run("rejects merge keys", func(t *testing.T) {
		t.Parallel()

		raw := `base: &base
  port: 8080
server:
  <<: *base
`
		_, err := manifest.Decode([]byte(raw), "merge.yaml")
		if err == nil {
			t.Fatal("expected error on merge keys, got nil")
		}
		if !errors.Is(err, errdef.ErrMergeKeyForbidden) {
			t.Errorf("got %v, want ErrMergeKeyForbidden", err)
		}
	})

	t.Run("rejects manifest exceeding size limit", func(t *testing.T) {
		t.Parallel()

		huge := make([]byte, manifest.MaxManifestSize+1)
		_, err := manifest.Decode(huge, "huge.yaml")
		if err == nil {
			t.Fatal("expected error on manifest exceeding 4 MiB, got nil")
		}
		if !errors.Is(err, errdef.ErrManifestTooLarge) {
			t.Errorf("got %v, want ErrManifestTooLarge", err)
		}
	})

	t.Run("rejects custom yaml tags", func(t *testing.T) {
		t.Parallel()

		raw := "key: !custom value\n"
		_, err := manifest.Decode([]byte(raw), "custom.yaml")
		if err == nil {
			t.Fatal("expected error on custom yaml tags, got nil")
		}
		if !errors.Is(err, errdef.ErrCustomTagForbidden) {
			t.Errorf("got %v, want ErrCustomTagForbidden", err)
		}
	})
}
