package catl

import (
	"errors"

	"github.com/ju4n97/catl/internal/manifest"
)

// ValidationError represents one or more schema or manifest validation failures.
type ValidationError = manifest.ValidationError

// Diagnostic represents an individual schema failure at a specific line and JSON pointer.
type Diagnostic = manifest.Diagnostic

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
	ErrRequiresNewer = errors.New("configuration requires a newer version of catl")
)
