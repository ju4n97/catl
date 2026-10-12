package config

import (
	"fmt"
	"regexp"
)

// identPattern matches valid catl identifiers: lowercase alphanumeric with single hyphens.
var identPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidateIdent checks that an identifier (catalog name, collection name, or ID)
// compiles with the identifier grammar: 1-63 bytes, lowercase ASCII alphanumerics
// segments separated by single hyphens.
func ValidateIdent(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidIdent)
	}
	if len(name) > 63 {
		return fmt.Errorf("%w: %q exceeds maximum length of 63 bytes", ErrInvalidIdent, name)
	}
	if !identPattern.MatchString(name) {
		return fmt.Errorf("%w: %q must match %q", ErrInvalidIdent, name, identPattern.String())
	}
	return nil
}
