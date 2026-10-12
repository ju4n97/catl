package config

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// ValidateRelativePath verifies that a relative path (used for schemas or collection dirs)
// is clean, uses forward slashes, doesn't escape the catalog root, has no dot-prefixed components,
// and is valid NFC-normalized UTF-8.
func ValidateRelativePath(p string) error {
	if p == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}
	if !utf8.ValidString(p) {
		return fmt.Errorf("%w: path %q is not valid UTF-8", ErrInvalidPath, p)
	}
	if !norm.NFC.IsNormalString(p) {
		return fmt.Errorf("%w: path %q must be Unicode NFC normalized", ErrInvalidPath, p)
	}
	if strings.Contains(p, "\\") {
		return fmt.Errorf("%w: path %q cannot contain backslashes", ErrInvalidPath, p)
	}
	for i := 0; i < len(p); i++ {
		c := p[i]
		if (c <= 0x1F) || c == 0x7F {
			return fmt.Errorf("%w: path %q contains ASCII control characters", ErrInvalidPath, p)
		}
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("%w: path %q cannot have a leading slash", ErrInvalidPath, p)
	}
	if strings.HasSuffix(p, "/") {
		return fmt.Errorf("%w: path %q cannot have a trailing slash", ErrInvalidPath, p)
	}
	if p == "." {
		return fmt.Errorf("%w: path cannot be '.'", ErrInvalidPath)
	}

	// Verify path components
	cleaned := path.Clean(p)
	if cleaned != p {
		return fmt.Errorf("%w: path %q must be normalized (want %q)", ErrInvalidPath, p, cleaned)
	}
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return fmt.Errorf("%w: path %q escapes root", ErrInvalidPath, p)
	}

	// Check individual segments for dot-prefixes or invalid component lengths
	for segment := range strings.SplitSeq(cleaned, "/") {
		if strings.HasPrefix(segment, ".") {
			return fmt.Errorf("%w: segment %q in %q cannot be dot-prefixed", ErrInvalidPath, segment, p)
		}
		if len(segment) > 255 {
			return fmt.Errorf("%w: segment %q exceeds 255 bytes", ErrInvalidPath, segment)
		}
	}

	return nil
}
