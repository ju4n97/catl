package manifest_test

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ju4n97/catl/internal/manifest"
	"github.com/ju4n97/catl/internal/scan"
)

// TestValidatorSchemaEnforcement verifies JSON Schema type and bound checks,
// line-number reporting, and permissive fallback when a schema is omitted.
func TestValidatorSchemaEnforcement(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"schemas/rules.json": &fstest.MapFile{
			Data: []byte(`{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "object",
				"required": ["enabled", "threshold"],
				"properties": {
					"enabled": { "type": "boolean" },
					"threshold": { "type": "integer", "maximum": 100 }
				},
				"additionalProperties": false
			}`),
		},
	}

	val := manifest.NewValidator(fsys)
	if err := val.Compile("schemas/rules.json"); err != nil {
		t.Fatalf("Compile() failed: %v", err)
	}

	t.Run("valid manifest against schema", func(t *testing.T) {
		t.Parallel()

		raw := "enabled: true\nthreshold: 50\n"
		m, err := manifest.Decode([]byte(raw), "rules/rate-limit.yaml")
		if err != nil {
			t.Fatalf("DecodeManifest() failed: %v", err)
		}

		ent := &scan.Entity{ID: "rate-limit", Shape: scan.EntityShapeFlat}
		if err := val.Validate(m, "schemas/rules.json", ent, nil); err != nil {
			t.Fatalf("ValidateManifest() unexpected error: %v", err)
		}
	})

	t.Run("value exceeding maximum bounds", func(t *testing.T) {
		t.Parallel()

		raw := "enabled: true\nthreshold: 150\n"
		m, err := manifest.Decode([]byte(raw), "rules/rate-limit.yaml")
		if err != nil {
			t.Fatalf("DecodeManifest() failed: %v", err)
		}

		ent := &scan.Entity{ID: "rate-limit", Shape: scan.EntityShapeFlat}
		err = val.Validate(m, "schemas/rules.json", ent, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}

		var valErr *manifest.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *schema.ValidationError, got %T", err)
		}

		formatted := valErr.Error()
		if !strings.Contains(
			formatted,
			"rules/rate-limit.yaml:2: /threshold: maximum: got 150, want 100",
		) {
			t.Errorf("diagnostic output mismatch: %s", formatted)
		}
	})

	t.Run("string boolean fails schema validation", func(t *testing.T) {
		t.Parallel()

		raw := "enabled: no\nthreshold: 50\n"
		m, err := manifest.Decode([]byte(raw), "rules/rate-limit.yaml")
		if err != nil {
			t.Fatalf("DecodeManifest() failed: %v", err)
		}

		ent := &scan.Entity{ID: "rate-limit", Shape: scan.EntityShapeFlat}
		err = val.Validate(m, "schemas/rules.json", ent, nil)
		if err == nil {
			t.Fatal("expected validation error on string boolean, got nil")
		}

		var valErr *manifest.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *schema.ValidationError, got %T", err)
		}

		formatted := valErr.Error()
		if !strings.Contains(
			formatted,
			"rules/rate-limit.yaml:1: /enabled: got string, want boolean",
		) {
			t.Errorf("diagnostic output mismatch: %s", formatted)
		}
	})

	t.Run("permissive mode passes when schema is omitted", func(t *testing.T) {
		t.Parallel()

		raw := "anything: allowed\nnumbers: 42\n"
		m, err := manifest.Decode([]byte(raw), "drafts/item.yaml")
		if err != nil {
			t.Fatalf("DecodeManifest() failed: %v", err)
		}

		ent := &scan.Entity{ID: "item", Shape: scan.EntityShapeFlat}
		if err := val.Validate(m, "", ent, nil); err != nil {
			t.Fatalf("permissive validation failed: %v", err)
		}
	})
}

// TestValidatorCatlFileFormat verifies the custom "catl-file" format assertion
// across bundle and flat entity types.
func TestValidatorCatlFileFormat(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"schemas/bundle.json": &fstest.MapFile{
			Data: []byte(`{
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type": "object",
				"required": ["runbook"],
				"properties": {
					"runbook": { "type": "string", "format": "catl-file" }
				}
			}`),
		},
	}

	val := manifest.NewValidator(fsys)
	if err := val.Compile("schemas/bundle.json"); err != nil {
		t.Fatalf("Compile() failed: %v", err)
	}

	t.Run("bundle entity with existing companion file passes", func(t *testing.T) {
		t.Parallel()

		bundleFS := fstest.MapFS{
			"notes.md": &fstest.MapFile{Data: []byte("# Notes")},
		}

		raw := "runbook: notes.md\n"
		m, err := manifest.Decode([]byte(raw), "services/billing/manifest.yaml")
		if err != nil {
			t.Fatalf("DecodeManifest() failed: %v", err)
		}

		ent := &scan.Entity{
			ID:        "billing",
			Shape:     scan.EntityShapeBundle,
			BundleDir: "services/billing",
		}
		if err := val.Validate(m, "schemas/bundle.json", ent, bundleFS); err != nil {
			t.Fatalf("ValidateManifest() failed: %v", err)
		}
	})

	t.Run("bundle entity with missing companion file fails", func(t *testing.T) {
		t.Parallel()

		bundleFS := fstest.MapFS{}

		raw := "runbook: missing.md\n"
		m, err := manifest.Decode([]byte(raw), "services/billing/manifest.yaml")
		if err != nil {
			t.Fatalf("DecodeManifest() failed: %v", err)
		}

		ent := &scan.Entity{
			ID:        "billing",
			Shape:     scan.EntityShapeBundle,
			BundleDir: "services/billing",
		}
		err = val.Validate(m, "schemas/bundle.json", ent, bundleFS)
		if err == nil {
			t.Fatal("expected error for missing companion file, got nil")
		}

		var valErr *manifest.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *schema.ValidationError, got %T", err)
		}

		if !strings.Contains(valErr.Error(), `file "missing.md" not found in bundle`) {
			t.Errorf("diagnostic output mismatch: %s", valErr.Error())
		}
	})

	t.Run("flat entity using catl-file format fails", func(t *testing.T) {
		t.Parallel()

		raw := "runbook: notes.md\n"
		m, err := manifest.Decode([]byte(raw), "services/auth.yaml")
		if err != nil {
			t.Fatalf("DecodeManifest() failed: %v", err)
		}

		ent := &scan.Entity{ID: "auth", Shape: scan.EntityShapeFlat}
		err = val.Validate(m, "schemas/bundle.json", ent, nil)
		if err == nil {
			t.Fatal("expected error when using catl-file on flat entity, got nil")
		}

		var valErr *manifest.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *schema.ValidationError, got %T", err)
		}

		if !strings.Contains(valErr.Error(), "catl-file format cannot be used on flat entities") {
			t.Errorf("diagnostic output mismatch: %s", valErr.Error())
		}
	})
}
