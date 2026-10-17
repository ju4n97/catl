// Package hash computes deterministic SHA-256 hashes for catalog content.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"

	"github.com/ju4n97/catl/internal/errdef"
)

// EmptyTreeHash is the SHA-256 digest of zero bytes with the canonical sha256: prefix.
const EmptyTreeHash = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// Tree walks fsys and computes the deterministic tree hash of all tracked catalog files.
// It excludes dot-prefixed paths and root catl.lock, rejects symlinks and control characters,
// sorts entries lexicographically, and computes the outer SHA-256 digest.
func Tree(fsys fs.FS) (string, error) {
	var paths []string

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Don't skip the root directory itself
		if path == "." {
			return nil
		}

		name := d.Name()

		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if path == "catl.lock" {
			return nil
		}

		// Abort on forbidden path characters
		for i := range len(path) {
			c := path[i]
			if c <= 0x1F || c == 0x7F || c == '\\' {
				return fmt.Errorf("%w: path %q contains control characters or backslashes", errdef.ErrInvalidPath, path)
			}
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%w: %s", errdef.ErrForbiddenSymlink, path)
		}

		if !d.IsDir() {
			paths = append(paths, path)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	// An empty catalog produces the SHA-256 of 0 bytes
	if len(paths) == 0 {
		return EmptyTreeHash, nil
	}

	// Sort paths in byte-wise lexicographical order (LC_ALL=C)
	slices.Sort(paths)

	h := sha256.New()
	for _, relPath := range paths {
		digest, err := File(fsys, relPath)
		if err != nil {
			return "", err
		}

		// Format: "<hex-digest>  <path>\n"
		if _, err := fmt.Fprintf(h, "%s  %s\n", digest, relPath); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// File computes the raw 64-character lowercase hexadecimal SHA-256 digest of a regular file.
func File(fsys fs.FS, path string) (string, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file %q: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read file %q: %w", path, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
