package regi

import "errors"

var (
	// ErrNotQualified is returned when an identifier is not qualified.
	ErrNotQualified = errors.New("regi: identifier is not qualified (want namespace.id)")

	// ErrNotFound is returned when an entity or namespace is not found.
	ErrNotFound = errors.New("regi: entity or namespace not found")

	// ErrLockStale is returned when the lock file is out of date.
	ErrLockStale = errors.New("regi: regi.lock is out of date; run 'regi update'")

	// ErrTamperedVendor is returned when the vendor directory content does not match the tree hash.
	ErrTamperedVendor = errors.New("regi: vendor directory content does not match tree hash")

	// ErrBlobNotCached is returned when a blob is not present in the local cache.
	ErrBlobNotCached = errors.New("regi: blob is not present in local cache (offline)")

	// ErrRequiresNewer is returned when a configuration requires a newer version of regi.
	ErrRequiresNewer = errors.New("regi: configuration requires a newer version of regi")
)
