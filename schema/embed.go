// Package schema embeds the JSON Schema Draft 2020-12 specification for catl.toml.
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

// compile parses and compiles raw JSON Schema Draft 2020-12 bytes.
func compile(url string, raw []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("schema: unmarshal %s: %w", url, err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err := compiler.AddResource(url, doc); err != nil {
		return nil, fmt.Errorf("schema: add resource %s: %w", url, err)
	}

	compiled, err := compiler.Compile(url)
	if err != nil {
		return nil, fmt.Errorf("schema: compile %s: %w", url, err)
	}
	return compiled, nil
}

var catlCompiled = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compile(CatlURL, CatlJSON)
})

// ValidateCatl validates generic decoded configuration data against catl.schema.json.
func ValidateCatl(data any) error {
	compiled, err := catlCompiled()
	if err != nil {
		return err
	}
	if err := compiled.Validate(data); err != nil {
		return fmt.Errorf("schema validation: %w", err)
	}
	return nil
}
