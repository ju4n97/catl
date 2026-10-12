package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ju4n97/catl/internal/config"
)

func TestValidateIdent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "single lowercase letter",
			input: "a",
		},
		{
			name:  "lowercase alphanumeric",
			input: "catalog123",
		},
		{
			name:  "hyphen-separated segments",
			input: "my-catalog-123",
		},
		{
			name:  "single digit",
			input: "1",
		},
		{
			name:  "maximum length",
			input: strings.Repeat("a", 63),
		},
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "exceeds maximum length",
			input:   strings.Repeat("a", 64),
			wantErr: true,
		},
		{
			name:    "uppercase letters",
			input:   "MyCatalog",
			wantErr: true,
		},
		{
			name:    "leading hyphen",
			input:   "-catalog",
			wantErr: true,
		},
		{
			name:    "trailing hyphen",
			input:   "catalog-",
			wantErr: true,
		},
		{
			name:    "consecutive hyphens",
			input:   "my--catalog",
			wantErr: true,
		},
		{
			name:    "underscore",
			input:   "my_catalog",
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   "my catalog",
			wantErr: true,
		},
		{
			name:    "non-ASCII characters",
			input:   "catalóg",
			wantErr: true,
		},
		{
			name:    "special characters",
			input:   "catalog!",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := config.ValidateIdent(tt.input)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Fatalf("ValidateIdent(%q) error = %v, wantErr %t", tt.input, err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, config.ErrInvalidIdent) {
				t.Errorf("ValidateIdent(%q) error = %v, want %v", tt.input, err, config.ErrInvalidIdent)
			}
		})
	}
}
