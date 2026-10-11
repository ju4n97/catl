package catl

import "errors"

var (
	// ErrNotQualified is returned when an identifier is not qualified.
	ErrNotQualified = errors.New("catl: identifier is not qualified (want namespace.id)")

	// ErrNotFound is returned when an entity or namespace is not found.
	ErrNotFound = errors.New("catl: entity or namespace not found")

	// ErrLockStale is returned when the lock file is out of date.
	ErrLockStale = errors.New("catl: catl.lock is out of date; run 'catl update'")

	// ErrTamperedVendor is returned when the vendor directory content does not match the tree hash.
	ErrTamperedVendor = errors.New("catl: vendor directory content does not match tree hash")

	// ErrRequiresNewer is returned when a configuration requires a newer version of catl.
	ErrRequiresNewer = errors.New("catl: configuration requires a newer version of catl")
)
