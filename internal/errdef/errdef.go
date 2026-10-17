// Package errdef defines the canonical sentinel errors and validation
// diagnostics shared by catl and its internal subsystems. Every package that
// needs to report or match one of these conditions imports this package
// directly; public packages re-export the values they promise to callers.
package errdef

import (
	"errors"
	"fmt"
	"strings"
)

// Diagnostic represents an individual schema failure at a specific line and JSON pointer.
type Diagnostic struct {
	Path    string `json:"path"`    // Manifest path relative to catalog root (e.g. "rules/rate-limit.yaml")
	Line    int    `json:"line"`    // 1-based source line number
	Pointer string `json:"pointer"` // JSON Pointer to the violation field (e.g. "/threshold")
	Message string `json:"message"` // Human-readable failure explanation
}

// String formats the diagnostic into standard compiler output: "path:line: pointer: message".
func (d Diagnostic) String() string {
	if d.Line > 0 {
		return fmt.Sprintf("%s:%d: %s: %s", d.Path, d.Line, d.Pointer, d.Message)
	}
	return fmt.Sprintf("%s: %s: %s", d.Path, d.Pointer, d.Message)
}

// ValidationError represents one or more schema or manifest validation failures.
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

var (
	// ErrNotQualified is returned when an identifier does not contain a namespace prefix.
	ErrNotQualified = errors.New("identifier is not qualified (want namespace.id)")

	// ErrNotFound is returned when a requested entity, collection, or namespace does not exist.
	ErrNotFound = errors.New("entity or namespace not found")

	// ErrLockStale indicates that catl.lock does not match declared dependencies in catl.toml.
	ErrLockStale = errors.New("catl.lock is out of date; run 'catl update'")

	// ErrTamperedVendor indicates that content in .catl/vendor does not match the recorded tree hash.
	ErrTamperedVendor = errors.New("vendor directory content does not match tree hash")

	// ErrRequiresNewer indicates that the configuration requires a newer version of the catl tool.
	ErrRequiresNewer = errors.New("catl.toml requires a newer version of catl")

	// ErrCRLFForbidden is returned when a structured file contains CRLF line endings.
	ErrCRLFForbidden = errors.New("CRLF line endings are forbidden (use LF)")

	// ErrMissing indicates a dependency's catalog files were not found on disk.
	ErrMissing = errors.New("dependency files not found")

	// ErrCorruptCache indicates cached dependency content does not match the recorded tree hash.
	ErrCorruptCache = errors.New("cached dependency content does not match tree hash")

	// ErrTargetMismatch indicates the resolved catalog's declared name does not match the dependency key.
	ErrTargetMismatch = errors.New("target catalog name does not match dependency key")

	// ErrForbiddenSymlink indicates a symbolic link was found in catalog content.
	ErrForbiddenSymlink = errors.New("symbolic links are forbidden")

	// ErrStrayRootEntry indicates an undeclared file or directory sits at the catalog root.
	ErrStrayRootEntry = errors.New("unrecognized file or directory at catalog root")

	// ErrInvalidIdent indicates an identifier violates the naming grammar.
	ErrInvalidIdent = errors.New("invalid identifier")

	// ErrInvalidPath indicates a path violates catalog layout invariants.
	ErrInvalidPath = errors.New("invalid path")

	// ErrConflictingEntity indicates an identifier is used as both a flat file and a bundle directory.
	ErrConflictingEntity = errors.New("conflicting entity shapes for identifier")

	// ErrMissingManifest indicates a bundle entity has no manifest.yaml or manifest.json.
	ErrMissingManifest = errors.New("bundle entity missing manifest.yaml or manifest.json")

	// ErrAmbiguousManifest indicates a bundle entity contains both manifest.yaml and manifest.json.
	ErrAmbiguousManifest = errors.New("bundle entity cannot contain both manifest.yaml and manifest.json")

	// ErrInvalidManifestType indicates an entity file does not use .yaml or .json.
	ErrInvalidManifestType = errors.New("manifest file must have .yaml or .json extension")

	// ErrManifestTooLarge indicates a manifest file exceeds 4 MiB.
	ErrManifestTooLarge = errors.New("manifest file exceeds 4 MiB limit")

	// ErrMultiDocument indicates a YAML stream contains multiple documents.
	ErrMultiDocument = errors.New("multi-document YAML streams are forbidden")

	// ErrMergeKeyForbidden indicates a YAML mapping contains merge keys ('<<').
	ErrMergeKeyForbidden = errors.New("YAML merge keys ('<<') are forbidden")

	// ErrCustomTagForbidden indicates a YAML node uses a custom tag.
	ErrCustomTagForbidden = errors.New("custom YAML tags are forbidden")

	// ErrInvalidLockfile indicates the lockfile contains syntax or schema violations.
	ErrInvalidLockfile = errors.New("invalid lockfile format")

	// ErrUnsupportedVersion indicates the lockfile format version is newer or unsupported.
	ErrUnsupportedVersion = errors.New("unsupported lockfile version")

	// ErrLockRequired indicates a remote dependency is missing its required lockfile pin.
	ErrLockRequired = errors.New("remote dependency requires lockfile entry")
)
