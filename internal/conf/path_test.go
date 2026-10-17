package conf_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/errdef"
)

// TestValidatePathComponent checks path component validation.
func TestValidatePathComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid components
		{name: "filename", input: "service.json"},
		{name: "directory", input: "schemas"},
		{name: "NFC Unicode", input: "caf\u00e9"},
		{name: "maximum length", input: strings.Repeat("a", 255)},

		// Invalid components
		{name: "empty", input: "", wantErr: true},
		{name: "exceeds length limit", input: strings.Repeat("a", 256), wantErr: true},
		{name: "forward slash", input: "schemas/service.json", wantErr: true},
		{name: "backslash", input: `schemas\service.json`, wantErr: true},
		{name: "null byte", input: "schema\x00.json", wantErr: true},
		{name: "control character", input: "schema\x1f.json", wantErr: true},
		{name: "delete character", input: "schema\x7f.json", wantErr: true},
		{name: "invalid UTF-8", input: "schema\xff.json", wantErr: true},
		{name: "non-NFC Unicode", input: "cafe\u0301", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := conf.ValidatePathComponent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ValidatePathComponent(%q) error = %v, wantErr %t",
					tt.input,
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr && !errors.Is(err, errdef.ErrInvalidPath) {
				t.Errorf(
					"ValidatePathComponent(%q) error = %v, want %v",
					tt.input,
					err,
					errdef.ErrInvalidPath,
				)
			}
		})
	}
}

// TestValidateRelativePath checks relative path validation.
func TestValidateRelativePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid relative paths
		{name: "single directory", input: "schemas"},
		{name: "nested file", input: "schemas/service.json"},
		{name: "leading current directory", input: "./schemas/service.json"},
		{name: "deeply nested path", input: "nested/dir/deep/file.json"},
		{name: "NFC Unicode", input: "caf\u00e9/manifest.yaml"},
		{name: "maximum component length", input: strings.Repeat("a", 255)},

		// Invalid path structure
		{name: "empty", input: "", wantErr: true},
		{name: "dot", input: ".", wantErr: true},
		{name: "current directory", input: "./", wantErr: true},
		{name: "repeated current directory", input: "././schemas/service.json", wantErr: true},
		{name: "root", input: "/", wantErr: true},
		{name: "leading slash", input: "/schemas", wantErr: true},
		{name: "trailing slash", input: "schemas/", wantErr: true},
		{name: "backslash", input: `schemas\service.json`, wantErr: true},
		{name: "dot segment", input: "schemas/./service.json", wantErr: true},
		{name: "parent segment", input: "schemas/sub/../service.json", wantErr: true},
		{name: "parent traversal", input: "../schemas", wantErr: true},
		{name: "parent directory", input: "..", wantErr: true},

		// Invalid path components
		{name: "dot-prefixed root component", input: ".schemas/service.json", wantErr: true},
		{name: "dot-prefixed nested component", input: "schemas/.service.json", wantErr: true},
		{name: "null byte", input: "schemas/\x00/service.json", wantErr: true},
		{name: "control character", input: "schemas/\x1f/service.json", wantErr: true},
		{name: "delete character", input: "schemas/\x7f/service.json", wantErr: true},
		{name: "invalid UTF-8", input: "schemas/\xff\xfe/service.json", wantErr: true},
		{name: "non-NFC Unicode", input: "cafe\u0301/manifest.yaml", wantErr: true},
		{name: "exceeds length limit", input: strings.Repeat("a", 256), wantErr: true},
		{
			name:    "nested component exceeds length limit",
			input:   "valid/" + strings.Repeat("b", 256),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := conf.ValidateRelativePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ValidateRelativePath(%q) error = %v, wantErr %t",
					tt.input,
					err,
					tt.wantErr,
				)
			}
			if tt.wantErr && !errors.Is(err, errdef.ErrInvalidPath) {
				t.Errorf(
					"ValidateRelativePath(%q) error = %v, want %v",
					tt.input,
					err,
					errdef.ErrInvalidPath,
				)
			}
		})
	}
}
