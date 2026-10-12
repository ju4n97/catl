package catl

import "errors"

var (
	// ErrNotQualified is returned when an identifier does not contain a namespace prefix.
	ErrNotQualified = errors.New("catl: identifier is not qualified (want namespace.id)")

	// ErrNotFound is returned when a requested entity, collection, or namespace does not exist.
	ErrNotFound = errors.New("catl: entity or namespace not found")

	// ErrLockStale indicates that catl.lock does not match declared dependencies in catl.toml.
	ErrLockStale = errors.New("catl: catl.lock is out of date; run 'catl update'")

	// ErrTamperedVendor indicates that content in .catl/vendor does not match the recorded tree hash.
	ErrTamperedVendor = errors.New("catl: vendor directory content does not match tree hash")

	// ErrRequiresNewer indicates that the configuration requires a newer version of the catl tool.
	ErrRequiresNewer = errors.New("catl: configuration requires a newer version of catl")
)
