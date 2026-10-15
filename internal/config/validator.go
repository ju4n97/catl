package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ju4n97/catl/schema"
)

const schemaURL = "https://raw.githubusercontent.com/ju4n97/catl/refs/heads/main/schema/catl.schema.json"

// getCompiledSchema parses and compiles catl.schema.json exactly once.
var getCompiledSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema.CatlJSON))
	if err != nil {
		return nil, fmt.Errorf("unmarshal catl schema: %w", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaURL, doc); err != nil {
		return nil, fmt.Errorf("register catl schema resource: %w", err)
	}

	compiled, err := c.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("compile catl schema: %w", err)
	}
	return compiled, nil
})

// ValidateSchema evaluates generic decoded TOML/JSON data against the embedded
// JSON Schema Draft 2020-12 specification.
func ValidateSchema(data any) error {
	compiled, err := getCompiledSchema()
	if err != nil {
		return err
	}
	if err := compiled.Validate(data); err != nil {
		return fmt.Errorf("schema validation: %w", err)
	}
	return nil
}

// Validate checks relational semantic constraints and filesystem path invariants
// on the File struct that cannot be verified by JSON Schema.
func (f *File) Validate() error {
	seenDirs := make(map[string]string, len(f.Collections))

	for name, coll := range f.Collections {
		// Deep OS path validation for schema file (if specified)
		if coll.Schema != "" {
			if err := ValidateRelativePath(coll.Schema); err != nil {
				return fmt.Errorf("collection %q schema: %w", name, err)
			}
		}

		// Resolve effective directory: defaults to collection name
		dir := coll.Dir
		if dir == "" {
			dir = name
		} else {
			if err := ValidateRelativePath(dir); err != nil {
				return fmt.Errorf("collection %q dir: %w", name, err)
			}
		}

		// Relational collision check: no two collections can share a target directory
		if existing, exists := seenDirs[dir]; exists {
			return fmt.Errorf(
				"collections %q and %q share the same directory %q",
				existing,
				name,
				dir,
			)
		}
		seenDirs[dir] = name
	}

	return nil
}
