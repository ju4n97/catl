package catl

import "github.com/ju4n97/catl/internal/errdef"

// ValidationError represents one or more schema or manifest validation failures.
type ValidationError = errdef.ValidationError

// Diagnostic represents an individual schema failure at a specific line and JSON pointer.
type Diagnostic = errdef.Diagnostic

var (
	// ErrNotQualified is returned when an identifier does not contain a namespace prefix.
	ErrNotQualified = errdef.ErrNotQualified

	// ErrNotFound is returned when a requested entity, collection, or namespace does not exist.
	ErrNotFound = errdef.ErrNotFound

	// ErrLockStale indicates that catl.lock does not match declared dependencies in catl.toml.
	ErrLockStale = errdef.ErrLockStale

	// ErrTamperedVendor indicates that content in .catl/vendor does not match the recorded tree hash.
	ErrTamperedVendor = errdef.ErrTamperedVendor

	// ErrRequiresNewer indicates that the configuration requires a newer version of the catl tool.
	ErrRequiresNewer = errdef.ErrRequiresNewer

	// ErrCRLFForbidden is returned when a structured file uses CRLF instead of LF.
	ErrCRLFForbidden = errdef.ErrCRLFForbidden

	// ErrMissing is returned when a dependency's catalog files are absent.
	ErrMissing = errdef.ErrMissing

	// ErrCorruptCache is returned when cached dependency content does not match its recorded tree hash.
	ErrCorruptCache = errdef.ErrCorruptCache

	// ErrTargetMismatch is returned when a dependency key does not match the target catalog's declared name.
	ErrTargetMismatch = errdef.ErrTargetMismatch

	// ErrForbiddenSymlink is returned when a symlink is found in catalog content.
	ErrForbiddenSymlink = errdef.ErrForbiddenSymlink

	// ErrStrayRootEntry is returned when an undeclared file or directory sits at the catalog root.
	ErrStrayRootEntry = errdef.ErrStrayRootEntry
)
