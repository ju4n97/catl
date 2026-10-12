package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ju4n97/catl/internal/config"
)

func TestValidateRelativePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "single directory segment",
			input: "schemas",
		},
		{
			name:  "nested file path",
			input: "schemas/service.json",
		},
		{
			name:  "deeply nested path",
			input: "nested/dir/deep/file.json",
		},
		{
			name:  "nfc normalized unicode path",
			input: "caf\u00e9/manifest.yaml",
		},
		{
			name:  "maximum allowed segment length of 255 bytes",
			input: strings.Repeat("a", 255),
		},
		{
			name:    "empty path",
			input:   "",
			wantErr: true,
		},
		{
			name:    "dot only",
			input:   ".",
			wantErr: true,
		},
		{
			name:    "slash only",
			input:   "/",
			wantErr: true,
		},
		{
			name:    "leading slash",
			input:   "/schemas",
			wantErr: true,
		},
		{
			name:    "trailing slash",
			input:   "schemas/",
			wantErr: true,
		},
		{
			name:    "contains backslash",
			input:   "schemas\\service.json",
			wantErr: true,
		},
		{
			name:    "unclean relative dot segment",
			input:   "schemas/./service.json",
			wantErr: true,
		},
		{
			name:    "unclean relative parent segment",
			input:   "schemas/sub/../service.json",
			wantErr: true,
		},
		{
			name:    "escapes root with parent traversal",
			input:   "../schemas",
			wantErr: true,
		},
		{
			name:    "escapes root directly",
			input:   "..",
			wantErr: true,
		},
		{
			name:    "root segment is dot prefixed",
			input:   ".schemas/service.json",
			wantErr: true,
		},
		{
			name:    "nested segment is dot prefixed",
			input:   "schemas/.service.json",
			wantErr: true,
		},
		{
			name:    "null byte control character",
			input:   "schemas/\x00/service.json",
			wantErr: true,
		},
		{
			name:    "ascii control character 0x1f",
			input:   "schemas/\x1f/service.json",
			wantErr: true,
		},
		{
			name:    "ascii delete character 0x7f",
			input:   "schemas/\x7f/service.json",
			wantErr: true,
		},
		{
			name:    "invalid utf8 byte sequence",
			input:   "schemas/\xff\xfe/service.json",
			wantErr: true,
		},
		{
			name:    "nfd unicode instead of nfc",
			input:   "cafe\u0301/manifest.yaml",
			wantErr: true,
		},
		{
			name:    "segment exceeds 255 bytes",
			input:   strings.Repeat("a", 256),
			wantErr: true,
		},
		{
			name:    "nested segment exceeds 255 bytes",
			input:   "valid/" + strings.Repeat("b", 256),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateRelativePath(tt.input)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("ValidateRelativePath(%q) error = %v, wantErr %t", tt.input, err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, config.ErrInvalidPath) {
				t.Errorf("ValidateRelativePath(%q) error = %v, want %v", tt.input, err, config.ErrInvalidPath)
			}
		})
	}
}
