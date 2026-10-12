package schema

import (
	"fmt"
	"strings"
)

// Diagnostic represents a single schema or manifest validation failure at a source location.
type Diagnostic struct {
	Path    string // Manifest path relative to catalog root (e.g. "rules/rate-limit.yaml")
	Line    int    // 1-based source line number
	Pointer string // JSON Pointer to the violation field (e.g. "/threshold")
	Message string // Human-readable failure explanation
}

// String formats the diagnostic into standard compiler output: "path:line: pointer: message".
func (d Diagnostic) String() string {
	if d.Line > 0 {
		return fmt.Sprintf("%s:%d: %s: %s", d.Path, d.Line, d.Pointer, d.Message)
	}
	return fmt.Sprintf("%s: %s: %s", d.Path, d.Pointer, d.Message)
}

// ValidationError aggregates one or more schema diagnostics across manifests.
type ValidationError struct {
	Diagnostics []Diagnostic
}

// Error formats all diagnostics into a newline-separated string.
func (e ValidationError) Error() string {
	if len(e.Diagnostics) == 0 {
		return "schema validation failed"
	}

	var sb strings.Builder
	for i, d := range e.Diagnostics {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(d.String())
	}
	return sb.String()
}
