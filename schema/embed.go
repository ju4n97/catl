// Package schema embeds and compiles JSON Schema Draft 2020-12 specifications.
package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// CatlURL is the canonical $id of the embedded catl.toml configuration schema.
const CatlURL = "https://raw.githubusercontent.com/ju4n97/catl/refs/heads/main/schema/catl.schema.json"

// CatlJSON contains the raw bytes of catl.schema.json.
//
//go:embed catl.schema.json
var CatlJSON []byte

// Compile parses and compiles raw JSON Schema Draft 2020-12 bytes into an executable Schema.
func Compile(url string, raw []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("schema: unmarshal %s: %w", url, err)
	}

	c := jsonschema.NewCompiler()
	c.AssertFormat()

	if err := c.AddResource(url, doc); err != nil {
		return nil, fmt.Errorf("schema: add resource %s: %w", url, err)
	}

	sch, err := c.Compile(url)
	if err != nil {
		return nil, fmt.Errorf("schema: compile %s: %w", url, err)
	}

	return sch, nil
}

// CatlCompiled returns the thread-safe singleton compiled Schema for catl.toml.
var CatlCompiled = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return Compile(CatlURL, CatlJSON)
})

// ValidateCatl validates generic decoded data against catl.schema.json.
func ValidateCatl(data any) error {
	sch, err := CatlCompiled()
	if err != nil {
		return err
	}
	if err := sch.Validate(data); err != nil {
		return fmt.Errorf("schema validation: %w", err)
	}
	return nil
}
