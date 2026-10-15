package lock

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/ju4n97/catl/internal/conf"
)

// CurrentVersion is the authoritative lockfile format version.
const CurrentVersion = 1

var (
	// ErrLockStale indicates that catl.lock does not match declared dependencies in catl.toml.
	ErrLockStale = errors.New("catl.lock is out of date; run 'catl update'")

	// ErrInvalidLockfile indicates that the lockfile contains syntax or schema violations.
	ErrInvalidLockfile = errors.New("invalid lockfile format")

	// ErrUnsupportedVersion indicates the lockfile format version is newer or unsupported.
	ErrUnsupportedVersion = errors.New("unsupported lockfile version")
)

// File models the root structure of a catl.lock file.
type File struct {
	Version      int     `toml:"version"    json:"version"`
	Dependencies []Entry `toml:"dependency" json:"dependencies,omitzero"`
}

// Entry represents an immutable pin for an individual remote dependency.
type Entry struct {
	Requested string `toml:"requested" json:"requested"`
	Resolved  string `toml:"resolved"  json:"resolved"`
	Tree      string `toml:"tree"      json:"tree"`
}

// Parse decodes and strictly validates raw TOML bytes into a File instance.
func Parse(data []byte) (*File, error) {
	if bytes.HasPrefix(data, []byte("\xef\xbb\xbf")) {
		return nil, fmt.Errorf("%w: UTF-8 BOM is forbidden", ErrInvalidLockfile)
	}

	var f File
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidLockfile, err.Error())
	}

	if err := f.Validate(); err != nil {
		return nil, err
	}

	return &f, nil
}

// Load reads and parses a catl.lock file from the local filesystem.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read lockfile %q: %w", path, err)
	}
	return Parse(data)
}

// LoadFS reads and parses a catl.lock file from an abstract fs.FS filesystem.
func LoadFS(fsys fs.FS, path string) (*File, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("read lockfile %q: %w", path, err)
	}
	return Parse(data)
}

// Marshal serializes a File into deterministic, sorted TOML bytes.
func Marshal(f *File) ([]byte, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}

	// Ensure dependencies are deterministically sorted by requested string
	sorted := slices.Clone(f.Dependencies)
	slices.SortFunc(sorted, func(a, b Entry) int {
		return strings.Compare(a.Requested, b.Requested)
	})

	toEncode := File{
		Version:      f.Version,
		Dependencies: sorted,
	}

	return toml.Marshal(toEncode)
}

// Save atomically writes a File to disk using a temporary file and rename.
func Save(targetPath string, f *File) error {
	data, err := Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}

	dir := filepath.Dir(targetPath)
	tmpFile, err := os.CreateTemp(dir, "catl.lock.*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary lockfile: %w", err)
	}
	tmpName := tmpFile.Name()

	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temporary lockfile: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temporary lockfile: %w", err)
	}

	if err := os.Rename(tmpName, targetPath); err != nil {
		return fmt.Errorf("atomic rename lockfile: %w", err)
	}

	return nil
}

// VerifyFreshness asserts that catl.lock strictly matches the declared dependencies in catl.toml.
// Returns ErrLockStale if any remote dependency is missing, extraneous, or mismatching.
func VerifyFreshness(declared map[string]conf.Dependency, lock *File) error {
	expected := make(map[string]string, len(declared))

	for depName, dep := range declared {
		canonical, err := CanonicalRequested(dep)
		if err != nil {
			return fmt.Errorf("verify freshness: %w", err)
		}
		if canonical != "" {
			expected[canonical] = depName
		}
	}

	remoteCount := len(expected)

	// If no remote dependencies are declared, lockfile must either be omitted or empty
	if remoteCount == 0 {
		if lock != nil && len(lock.Dependencies) > 0 {
			return fmt.Errorf(
				"%w: catl.lock contains %d pins but catl.toml declares no remote dependencies",
				ErrLockStale,
				len(lock.Dependencies),
			)
		}
		return nil
	}

	// Remote dependencies are declared, so a valid lockfile is strictly required
	if lock == nil {
		return fmt.Errorf("%w: catalog declares remote dependencies but catl.lock is missing", ErrLockStale)
	}

	locked := make(map[string]string, len(lock.Dependencies))
	for _, entry := range lock.Dependencies {
		locked[entry.Requested] = entry.Resolved
	}

	// Verify every declared remote dependency has an exact pin in the lockfile
	for canonical, depName := range expected {
		if _, exists := locked[canonical]; !exists {
			return fmt.Errorf("%w: missing pin for dependency %q (%s)", ErrLockStale, depName, canonical)
		}
	}

	// Verify there are no extraneous pins from deleted dependencies
	if len(locked) != remoteCount {
		return fmt.Errorf(
			"%w: catl.lock contains %d pins but catl.toml declares %d remote dependencies",
			ErrLockStale,
			len(locked),
			remoteCount,
		)
	}

	return nil
}

// CanonicalRequested derives the unique 'requested' string used as the key in catl.lock.
// Returns an empty string and nil error for path dependencies (which are never locked).
func CanonicalRequested(dep conf.Dependency) (string, error) {
	src, err := dep.Source()
	if err != nil {
		return "", err
	}

	switch s := src.(type) {
	case conf.GitSource:
		var ref string
		switch {
		case s.Tag != "":
			ref = "tag=" + url.QueryEscape(s.Tag)
		case s.Branch != "":
			ref = "branch=" + url.QueryEscape(s.Branch)
		case s.Rev != "":
			ref = "rev=" + s.Rev
		}

		req := "git+" + s.Repo + "#" + ref
		if s.Subdir != "" {
			req += "&subdir=" + s.Subdir
		}
		return req, nil

	case conf.PathSource:
		return "", nil

	default:
		return "", fmt.Errorf("unsupported source type %T", src)
	}
}

// Find retrieves an entry matching the requested canonical source string.
func (f *File) Find(requested string) (Entry, bool) {
	for _, entry := range f.Dependencies {
		if entry.Requested == requested {
			return entry, true
		}
	}
	return Entry{}, false
}

// Validate checks structural constraints, version numbers, and hash formats on f.
func (f *File) Validate() error {
	if f.Version == 0 {
		return fmt.Errorf("%w: version is required", ErrInvalidLockfile)
	}
	if f.Version != CurrentVersion {
		return fmt.Errorf("%w: got version %d, want %d", ErrUnsupportedVersion, f.Version, CurrentVersion)
	}

	seen := make(map[string]struct{}, len(f.Dependencies))
	for _, entry := range f.Dependencies {
		if entry.Requested == "" {
			return fmt.Errorf("%w: entry requested string cannot be empty", ErrInvalidLockfile)
		}
		if entry.Resolved == "" {
			return fmt.Errorf("%w: entry resolved string cannot be empty for %q", ErrInvalidLockfile, entry.Requested)
		}
		if err := validateGitCommit(entry.Resolved); err != nil {
			return fmt.Errorf("%w: invalid resolved commit for %q: %s", ErrInvalidLockfile, entry.Requested, err.Error())
		}
		if err := validateTreeHash(entry.Tree); err != nil {
			return fmt.Errorf("%w: invalid tree hash for %q: %s", ErrInvalidLockfile, entry.Requested, err.Error())
		}

		if _, exists := seen[entry.Requested]; exists {
			return fmt.Errorf("%w: duplicate entry for %q", ErrInvalidLockfile, entry.Requested)
		}
		seen[entry.Requested] = struct{}{}
	}

	return nil
}

// validateGitCommit checks that sha is a full 40-character lowercase hexadecimal
// Git hash. Full hashes avoid ambiguity when multiple commits share the same prefix.
func validateGitCommit(sha string) error {
	if len(sha) != 40 {
		return fmt.Errorf("resolved commit must be a 40-character hex string (got %d chars)", len(sha))
	}

	for i := range len(sha) {
		c := sha[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("contains non-hex character %q at index %d", c, i)
		}
	}

	return nil
}

// validateTreeHash checks that a tree hash matches the required sha256:[0-9a-f]{64} format.
func validateTreeHash(h string) error {
	if !strings.HasPrefix(h, "sha256:") || len(h) != 71 {
		return errors.New("must be formatted as 'sha256:' followed by 64 hex characters")
	}

	for i := 7; i < len(h); i++ {
		c := h[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("contains non-lowercase-hex character %q at index %d", c, i)
		}
	}

	return nil
}
