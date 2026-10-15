package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ju4n97/catl/internal/scan"
)

// Validator compiles collection schemas and validates entity manifests against them.
type Validator struct {
	fsys       fs.FS
	schemas    map[string]*jsonschema.Schema
	validateMu sync.Mutex
	activeCtx  *validationContext
}

// validationContext supplies per-manifest state required during custom format assertion callbacks.
type validationContext struct {
	entity   *scan.Entity
	bundleFS fs.FS
}

// NewValidator initializes a Validator backed by a catalog filesystem.
func NewValidator(fsys fs.FS) *Validator {
	return &Validator{
		fsys:    fsys,
		schemas: make(map[string]*jsonschema.Schema),
	}
}

// Compile compiles a JSON Schema Draft 2020-12 file for a collection.
func (v *Validator) Compile(schemaPath string) error {
	if schemaPath == "" {
		return nil
	}

	if _, exists := v.schemas[schemaPath]; exists {
		return nil
	}

	data, err := fs.ReadFile(v.fsys, schemaPath)
	if err != nil {
		return fmt.Errorf("read schema %s: %w", schemaPath, err)
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("unmarshal schema %s: %w", schemaPath, err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()

	compiler.RegisterFormat(&jsonschema.Format{
		Name: "catl-file",
		Validate: func(val any) error {
			return v.validateCatlFile(val)
		},
	})

	canonicalURL := "catl://" + schemaPath
	if err := compiler.AddResource(canonicalURL, doc); err != nil {
		return fmt.Errorf("register schema %q: %w", schemaPath, err)
	}

	sch, err := compiler.Compile(canonicalURL)
	if err != nil {
		return fmt.Errorf("compile schema %s: %w", schemaPath, err)
	}

	v.schemas[schemaPath] = sch
	return nil
}

// Validate validates a decoded manifest against its collection's schema.
func (v *Validator) Validate(
	m *Manifest,
	schemaPath string,
	entity *scan.Entity,
	bundleFS fs.FS,
) error {
	if schemaPath == "" {
		return nil
	}

	sch, ok := v.schemas[schemaPath]
	if !ok {
		if err := v.Compile(schemaPath); err != nil {
			return err
		}
		sch = v.schemas[schemaPath]
	}

	v.validateMu.Lock()
	v.activeCtx = &validationContext{entity: entity, bundleFS: bundleFS}
	err := sch.Validate(m.Raw)
	v.activeCtx = nil
	v.validateMu.Unlock()

	if err == nil {
		return nil
	}

	var schemaErr *jsonschema.ValidationError
	if !errors.As(err, &schemaErr) {
		return err
	}

	return v.formatValidationError(m, schemaErr)
}

// validateCatlFile checks that "catl-file" properties resolve to existing companion files
// inside a bundle entity and rejects usage on flat entities.
func (v *Validator) validateCatlFile(val any) error {
	s, ok := val.(string)
	if !ok {
		return nil
	}

	ctx := v.activeCtx
	if ctx == nil || ctx.entity == nil {
		return nil
	}

	if ctx.entity.Shape != scan.EntityShapeBundle {
		return errors.New("catl-file format cannot be used on flat entities")
	}

	if strings.Contains(s, "\\") {
		return fmt.Errorf("companion path %q cannot contain backslashes", s)
	}
	if strings.HasPrefix(s, "/") {
		return fmt.Errorf("companion path %q cannot have leading slashes", s)
	}
	if strings.HasPrefix(s, "../") || strings.Contains(s, "/../") || s == ".." {
		return fmt.Errorf("companion path %q cannot traverse with '..'", s)
	}

	cleaned := path.Clean(s)
	if cleaned != s {
		return fmt.Errorf("companion path %q is not clean", s)
	}

	if ctx.bundleFS != nil {
		if _, err := fs.Stat(ctx.bundleFS, s); err != nil {
			return fmt.Errorf("file %q not found in bundle", s)
		}
	}

	return nil
}

// formatValidationError flattens the jsonschema error hierarchy into structured Diagnostics.
func (v *Validator) formatValidationError(
	m *Manifest,
	schemaErr *jsonschema.ValidationError,
) error {
	unit := schemaErr.BasicOutput()
	var diags []Diagnostic

	var collect func(u *jsonschema.OutputUnit)
	collect = func(u *jsonschema.OutputUnit) {
		if u == nil {
			return
		}
		if u.Error != nil {
			ptr := u.InstanceLocation
			if ptr == "" {
				ptr = "/"
			}
			line := m.FindLine(ptr)

			diags = append(diags, Diagnostic{
				Path:    m.Path,
				Line:    line,
				Pointer: ptr,
				Message: u.Error.String(),
			})
		}
		for i := range u.Errors {
			collect(&u.Errors[i])
		}
	}

	collect(unit)

	if len(diags) == 0 {
		return &ValidationError{
			Diagnostics: []Diagnostic{
				{
					Path:    m.Path,
					Line:    1,
					Pointer: "/",
					Message: schemaErr.Error(),
				},
			},
		}
	}

	return &ValidationError{Diagnostics: diags}
}
