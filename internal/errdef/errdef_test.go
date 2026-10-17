package errdef_test

import (
	"encoding/json"
	"testing"

	"github.com/ju4n97/catl/internal/errdef"
)

// TestDiagnosticString verifies standard compiler-style formatting of individual diagnostics.
func TestDiagnosticString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		diag errdef.Diagnostic
		want string
	}{
		{
			name: "diagnostic with line number",
			diag: errdef.Diagnostic{
				Path:    "rules/auth.yaml",
				Line:    12,
				Pointer: "/threshold",
				Message: "value exceeds maximum",
			},
			want: "rules/auth.yaml:12: /threshold: value exceeds maximum",
		},
		{
			name: "diagnostic without line number",
			diag: errdef.Diagnostic{
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

// TestDiagnosticJSON verifies that marshaled diagnostics use the documented lowercase keys.
func TestDiagnosticJSON(t *testing.T) {
	t.Parallel()

	got, err := json.Marshal(errdef.Diagnostic{
		Path:    "a.yaml",
		Line:    2,
		Pointer: "/x",
		Message: "bad",
	})
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	want := `{"path":"a.yaml","line":2,"pointer":"/x","message":"bad"}`
	if string(got) != want {
		t.Errorf("json.Marshal() = %s, want %s", got, want)
	}
}

// TestValidationError verifies that ValidationError aggregates multiple diagnostics separated by newlines.
func TestValidationError(t *testing.T) {
	t.Parallel()

	t.Run("empty diagnostics fallback", func(t *testing.T) {
		t.Parallel()

		err := &errdef.ValidationError{}
		if got := err.Error(); got != "schema validation failed" {
			t.Errorf("Error() = %q, want 'schema validation failed'", got)
		}
	})

	t.Run("multiple diagnostics formatted across lines", func(t *testing.T) {
		t.Parallel()

		err := &errdef.ValidationError{
			Diagnostics: []errdef.Diagnostic{
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
