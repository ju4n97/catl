package conf

import "errors"

var (
	// ErrInvalidIdent is returned when an identifier violates the naming grammar.
	ErrInvalidIdent = errors.New("invalid identifier")

	// ErrInvalidPath indicates a configuration path violates catalog layout invariants.
	ErrInvalidPath = errors.New("invalid configuration path")

	// ErrRequiresNewer indicates that the configuration contains unrecognized keys
	// or tables, signaling that a newer version of catl is required.
	ErrRequiresNewer = errors.New("catl.toml requires a newer version of catl")
)
