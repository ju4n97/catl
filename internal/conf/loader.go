package conf

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"

	"github.com/ju4n97/catl/schema"
)

var utf8BOM = []byte("\xef\xbb\xbf")

// LoadFile reads a catl.toml file from disk and parses its contents.
func LoadFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	return Parse(data)
}

// Parse decodes raw TOML configuration bytes. It checks for a UTF-8 BOM, validates
// with strict field checking, executes JSON Schema validation, and runs relational validations.
func Parse(data []byte) (*File, error) {
	if bytes.HasPrefix(data, utf8BOM) {
		return nil, errors.New("catl.toml: UTF-8 BOM is forbidden")
	}

	if bytes.Contains(data, []byte("\r")) {
		return nil, errors.New("catl.toml: CRLF line endings are forbidden (use LF)")
	}

	var file File
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&file); err != nil {
		if strictErr, ok := errors.AsType[*toml.StrictMissingError](err); ok {
			return nil, fmt.Errorf("%w: %s", ErrRequiresNewer, strictErr.String())
		}
		return nil, fmt.Errorf("decode catl.toml: %w", err)
	}

	var generic any
	if err := toml.Unmarshal(data, &generic); err != nil {
		return nil, fmt.Errorf("parse toml: %w", err)
	}
	if err := schema.ValidateCatl(generic); err != nil {
		return nil, err
	}

	if err := file.ValidateCollections(); err != nil {
		return nil, err
	}

	return &file, nil
}

// Marshal serializes the configuration file into formatted TOML bytes.
func (f *File) Marshal() ([]byte, error) {
	if err := f.ValidateCollections(); err != nil {
		return nil, fmt.Errorf("validate config before marshaling: %w", err)
	}
	return toml.Marshal(f)
}
