package skel_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ju4n97/catl/internal/skel"
)

// TestCopy verifies template rendering, file placement, and placeholder substitution
// across catalog and workspace starters.
func TestCopy(t *testing.T) {
	t.Parallel()

	t.Run("render minimal catalog template", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		created, err := skel.Copy(skel.TemplateMinimal, tmpDir, skel.Data{
			Name:       "app-rules",
			Collection: "rules",
		})
		if err != nil {
			t.Fatalf("Copy() failed: %v", err)
		}

		if len(created) != 3 {
			t.Fatalf("expected 3 files, got %d: %v", len(created), created)
		}

		// Verify catl.toml content
		catlPath := filepath.Join(tmpDir, "catl.toml")
		catlData, err := os.ReadFile(catlPath)
		if err != nil {
			t.Fatalf("read catl.toml failed: %v", err)
		}

		wantTOMLPrefix := "#:schema https://ju4n97.github.io/catl/schema/v1/catl.schema.json"
		if !strings.Contains(string(catlData), wantTOMLPrefix) {
			t.Errorf("expected catl.toml to contain schema header")
		}
		if !strings.Contains(string(catlData), `name = "app-rules"`) {
			t.Errorf("expected catl.toml to contain name")
		}
		if !strings.Contains(string(catlData), `[collections.rules]`) {
			t.Errorf("expected catl.toml to contain collections.rules")
		}

		// Verify schemas/rules.json exists
		schemaPath := filepath.Join(tmpDir, "schemas", "rules.json")
		if _, err := os.Stat(schemaPath); err != nil {
			t.Errorf("expected %s to exist: %v", schemaPath, err)
		}

		// Verify .gitattributes exists
		gitAttrPath := filepath.Join(tmpDir, ".gitattributes")
		if _, err := os.Stat(gitAttrPath); err != nil {
			t.Errorf("expected %s to exist: %v", gitAttrPath, err)
		}
	})

	t.Run("render minimal workspace template", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		created, err := skel.Copy(skel.TemplateWorkspaceMinimal, tmpDir, skel.Data{})
		if err != nil {
			t.Fatalf("Copy() failed: %v", err)
		}

		if len(created) != 2 {
			t.Fatalf("expected 2 files, got %d: %v", len(created), created)
		}

		catlPath := filepath.Join(tmpDir, "catl.toml")
		catlData, err := os.ReadFile(catlPath)
		if err != nil {
			t.Fatalf("read workspace catl.toml failed: %v", err)
		}

		if !strings.Contains(string(catlData), "[workspace]\nmembers = []") {
			t.Errorf("workspace catl.toml missing empty members array")
		}

		// Workspaces must never contain schemas at the root
		schemaPath := filepath.Join(tmpDir, "schemas")
		if _, err := os.Stat(schemaPath); !os.IsNotExist(err) {
			t.Errorf("schemas directory should not exist in workspace root")
		}
	})

	t.Run("substitutes custom collection in filenames and content", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		_, err := skel.Copy(skel.TemplateMinimal, tmpDir, skel.Data{
			Name:       "billing-catalog",
			Collection: "invoices",
		})
		if err != nil {
			t.Fatalf("Copy() failed: %v", err)
		}

		schemaPath := filepath.Join(tmpDir, "schemas", "invoices.json")
		if _, err := os.Stat(schemaPath); err != nil {
			t.Errorf(
				"expected placeholder __collection__ to be replaced with invoices.json: %v",
				err,
			)
		}
	})

	t.Run("fails on unknown template", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		_, err := skel.Copy("non-existent-template", tmpDir, skel.Data{})
		if err == nil {
			t.Fatal("expected error on unknown template, got nil")
		}
	})
}
