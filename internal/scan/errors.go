package scan

import "errors"

var (
	// ErrStrayRootEntry is returned when an unrecognized non-dot file or directory
	// is found in the catalog root.
	ErrStrayRootEntry = errors.New("unrecognized file or directory at catalog root")

	// ErrForbiddenSymlink indicates a symbolic link was detected outside an ignored dot-prefixed directory.
	ErrForbiddenSymlink = errors.New("symbolic links are forbidden")

	// ErrConflictingEntity is returned when an identifier is used as both a flat file
	// and a bundle directory in the same collection.
	ErrConflictingEntity = errors.New("conflicting entity shapes for identifier")

	// ErrMissingManifest indicates a bundle entity directory does not contain manifest.yaml or manifest.json.
	ErrMissingManifest = errors.New("bundle entity missing manifest.yaml or manifest.json")

	// ErrAmbiguousManifest indicates a bundle entity contains both manifest.yaml and manifest.json.
	ErrAmbiguousManifest = errors.New("bundle entity cannot contain both manifest.yaml and manifest.json")

	// ErrInvalidManifestType is returned when an entity file does not use a supported extension (.yaml or .json).
	ErrInvalidManifestType = errors.New("manifest file must have .yaml or .json extension")
)
