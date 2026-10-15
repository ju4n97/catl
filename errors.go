package catl

import (
	"errors"

	"github.com/ju4n97/catl/internal/conf"
	"github.com/ju4n97/catl/internal/lock"
	"github.com/ju4n97/catl/internal/manifest"
	"github.com/ju4n97/catl/internal/source"
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
	ErrLockStale = lock.ErrLockStale

	// ErrTamperedVendor indicates that content in .catl/vendor does not match the recorded tree hash.
	ErrTamperedVendor = source.ErrTamperedVendor

	// ErrRequiresNewer indicates that the configuration requires a newer version of the catl tool.
	ErrRequiresNewer = conf.ErrRequiresNewer
)
