package schema_test

import (
	"testing"

	"github.com/ju4n97/catl/internal/schema"
)

// TestDiagnosticString verifies standard compiler-style formatting of individual diagnostics.
func TestDiagnosticString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		diag schema.Diagnostic
		want string
	}{
		{
			name: "diagnostic with line number",
			diag: schema.Diagnostic{
				Path:    "rules/auth.yaml",
				Line:    12,
				Pointer: "/threshold",
				Message: "value exceeds maximum",
			},
			want: "rules/auth.yaml:12: /threshold: value exceeds maximum",
		},
		{
			name: "diagnostic without line number",
			diag: schema.Diagnostic{
				Path:    "rules/auth.yaml",
				Line:    0,
				Pointer: "/threshold",
				Message: "value exceeds maximum",
			},
			want: "rules/auth.yaml: /threshold: value exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.diag.String(); got != tt.want {
				t.Errorf("Diagnostic.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestValidationError verifies that ValidationError aggregates multiple diagnostics separated by newlines.
func TestValidationError(t *testing.T) {
	t.Parallel()

	t.Run("empty diagnostics fallback", func(t *testing.T) {
		t.Parallel()

		err := &schema.ValidationError{}
		if got := err.Error(); got != "schema validation failed" {
			t.Errorf("Error() = %q, want 'schema validation failed'", got)
		}
	})

	t.Run("multiple diagnostics formatted across lines", func(t *testing.T) {
		t.Parallel()

		err := &schema.ValidationError{
			Diagnostics: []schema.Diagnostic{
				{Path: "a.yaml", Line: 1, Pointer: "/p1", Message: "err1"},
				{Path: "b.yaml", Line: 2, Pointer: "/p2", Message: "err2"},
			},
		}
		want := "a.yaml:1: /p1: err1\nb.yaml:2: /p2: err2"
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}
