package conf

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// ValidatePathComponent checks that an individual directory or file name
// is valid UTF-8, NFC-normalized, does not exceed 255 bytes, and contains no
// slashes, backslashes, or ASCII control characters.
func ValidatePathComponent(name string) error {
	if name == "" {
		return fmt.Errorf("%w: path component cannot be empty", ErrInvalidPath)
	}
	if len(name) > 255 {
		return fmt.Errorf("%w: path component %q exceeds 255 bytes", ErrInvalidPath, name)
	}
	if !utf8.ValidString(name) {
		return fmt.Errorf("%w: path component %q is not valid UTF-8", ErrInvalidPath, name)
	}
	if !norm.NFC.IsNormalString(name) {
		return fmt.Errorf(
			"%w: path component %q must be Unicode NFC normalized",
			ErrInvalidPath,
			name,
		)
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf(
			"%w: path component %q cannot contain slashes or backslashes",
			ErrInvalidPath,
			name,
		)
	}
	for i := range len(name) {
		c := name[i]
		if (c <= 0x1F) || c == 0x7F {
			return fmt.Errorf(
				"%w: path component %q contains ASCII control characters",
				ErrInvalidPath,
				name,
			)
		}
	}
	return nil
}

// ValidateRelativePath verifies that a relative path (used for schemas or collection dirs)
// is clean, uses forward slashes, doesn't escape the catalog root, has no dot-prefixed components,
// and consists exclusively of valid path components.
func ValidateRelativePath(p string) error {
	if p == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
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

	cleaned := path.Clean(p)
	if cleaned != p {
		return fmt.Errorf("%w: path %q must be normalized (want %q)", ErrInvalidPath, p, cleaned)
	}
	if strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return fmt.Errorf("%w: path %q escapes root", ErrInvalidPath, p)
	}

	for segment := range strings.SplitSeq(cleaned, "/") {
		if strings.HasPrefix(segment, ".") {
			return fmt.Errorf(
				"%w: segment %q in %q cannot be dot-prefixed",
				ErrInvalidPath,
				segment,
				p,
			)
		}
		if err := ValidatePathComponent(segment); err != nil {
			return err
		}
	}

	return nil
}
