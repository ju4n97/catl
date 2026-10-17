// Package manifest decodes and validates entity manifests.
package manifest

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"

	"github.com/ju4n97/catl/internal/errdef"
)

const (
	// MaxManifestSize enforces the catalog limit of 4 MiB per manifest file.
	MaxManifestSize = 4 * 1024 * 1024
)

// Manifest holds decoded manifest data along with its parsed AST for line lookups.
type Manifest struct {
	Path string
	Raw  any
	File *ast.File
}

// Decode parses raw YAML or JSON bytes enforcing YAML 1.2 core schema rules,
// unique keys, single document streams, and limits.
func Decode(data []byte, relPath string) (*Manifest, error) {
	if len(data) > MaxManifestSize {
		return nil, fmt.Errorf("%w: %s (%d bytes)", errdef.ErrManifestTooLarge, relPath, len(data))
	}

	if bytes.Contains(data, []byte("\r")) {
		return nil, fmt.Errorf("%s: %w", relPath, errdef.ErrCRLFForbidden)
	}

	// Parse AST to verify document boundaries and token line positions
	file, err := parser.ParseBytes(data, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", relPath, err)
	}

	if len(file.Docs) == 0 {
		return &Manifest{
			Path: relPath,
			Raw:  map[string]any{},
			File: file,
		}, nil
	}
	if len(file.Docs) > 1 {
		return nil, fmt.Errorf(
			"%w: %s contains %d documents",
			errdef.ErrMultiDocument,
			relPath,
			len(file.Docs),
		)
	}

	// Reject merge keys ('<<') and custom tags ('!') across the AST
	if err := inspectAST(file.Docs[0].Body); err != nil {
		return nil, fmt.Errorf("%s: %w", relPath, err)
	}

	// Decode into generic Go types with strict duplicate key checking
	// Also, in YAML 1.2 Core Schema, 'no', 'yes', 'on', 'off' decode as strings, not booleans
	var raw any
	if err := yaml.UnmarshalWithOptions(data, &raw, yaml.Strict()); err != nil {
		return nil, fmt.Errorf("decode manifest %s: %w", relPath, err)
	}

	return &Manifest{
		Path: relPath,
		Raw:  raw,
		File: file,
	}, nil
}

// FindLine maps a JSON Pointer (e.g. "/threshold" or "/items/0") to its 1-based source line number.
func (m *Manifest) FindLine(pointer string) int {
	if m.File == nil || len(m.File.Docs) == 0 || m.File.Docs[0].Body == nil {
		return 1
	}

	curr := m.File.Docs[0].Body
	lastLine := curr.GetToken().Position.Line

	for part := range strings.SplitSeq(strings.TrimPrefix(pointer, "/"), "/") {
		if part == "" {
			continue
		}

		switch n := curr.(type) {
		case *ast.MappingNode:
			var next ast.Node
			var matchedLine int
			for _, entry := range n.Values {
				if entry.Key != nil && entry.Key.GetToken() != nil &&
					entry.Key.GetToken().Value == part {
					matchedLine = entry.Key.GetToken().Position.Line
					next = entry.Value
					break
				}
			}
			if next == nil {
				return lastLine
			}
			lastLine = matchedLine
			curr = next

		case *ast.SequenceNode:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(n.Values) {
				return lastLine
			}
			curr = n.Values[idx]
			if curr != nil && curr.GetToken() != nil {
				lastLine = curr.GetToken().Position.Line
			}

		default:
			return lastLine
		}
	}

	return lastLine
}

// inspectAST recursively traverses the AST to reject merge keys and custom YAML tags.
func inspectAST(node ast.Node) error {
	if node == nil {
		return nil
	}

	if node.Type() == ast.TagType {
		return fmt.Errorf("%w: %q", errdef.ErrCustomTagForbidden, node.GetToken().Value)
	}
	if node.Type() == ast.MergeKeyType {
		return errdef.ErrMergeKeyForbidden
	}

	switch n := node.(type) {
	case *ast.MappingNode:
		for _, entry := range n.Values {
			if entry.Key != nil && entry.Key.Type() == ast.MergeKeyType {
				return errdef.ErrMergeKeyForbidden
			}
			if err := inspectAST(entry.Key); err != nil {
				return err
			}
			if err := inspectAST(entry.Value); err != nil {
				return err
			}
		}
	case *ast.SequenceNode:
		for _, val := range n.Values {
			if err := inspectAST(val); err != nil {
				return err
			}
		}
	}

	return nil
}
