package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/ju4n97/catl/schema"
)

const schemaURL = "https://ju4n97.github.io/catl/schema/v1/catl.schema.json"

// getCompiledSchema parses and compiles catl.schema.json exactly once.
var getCompiledSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema.CatlBytes))
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

// Validate checks post-schema semantic constraints on the File struct, specifically
// verifying that collection directories don't overlap or collide.
func (f *File) Validate() error {
	if f.Workspace != nil {
		return nil
	}

	seenDirs := make(map[string]string, len(f.Collections))
	for name, coll := range f.Collections {
		dir := coll.Dir
		if dir == "" {
			dir = name
		}

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
