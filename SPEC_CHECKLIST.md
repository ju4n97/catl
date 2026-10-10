# Specification Verification Matrix

## 1. Catalog & Layout

- [ ] Missing default collection directory treated as empty collection (no .gitkeep required).
- [ ] Catalog root rejects non-dot stray files (clean checkout rule).
- [ ] Path names with control chars (\x00-\x1F, \x7F) or backslashes (\) rejected globally.
- [ ] Symlinks anywhere in catalog outside dot-paths rejected.
- [ ] Flat entity cannot have companion files.
- [ ] Bundle entity must contain exactly one manifest (yaml or json).
- [ ] ID cannot collide as both flat file and bundle directory.

## 2. Validation & Schemas

- [ ] Missing schema in regi.toml defaults to permissive schema `{}`.
- [ ] Manifest YAML 1.2 core parsing: `on`/`off`/`yes`/`no` are strings, not booleans.
- [ ] Duplicate YAML/JSON keys rejected.
- [ ] `format: regi-file` asserts file exists in bundle (inline or .blob).
- [ ] `format: regi-file` fails if declared on a flat entity.

## 3. Storage & Blobs

- [ ] `.blob` file size capped at 4 KiB.
- [ ] `video.mp4` and `video.mp4.blob` coexisting is a validation error.
- [ ] `regi vendor` copies referenced blobs to `.regi/vendor/blobs/`.
- [ ] `regi vendor --no-blobs` skips blob downloads.
- [ ] Reading an uncached blob while `--offline` returns `ErrBlobNotCached`.
- [ ] Blob download aborted if incoming bytes exceed `size`.
- [ ] Mismatched blob digest rejected; cache entry is not written.

## 4. Integrity & Tree Hash

- [ ] Tree hash excludes root `regi.lock`.
- [ ] Tree hash excludes dot-prefixed paths.
- [ ] Tree hash reproducible with documented shell recipe.
- [ ] `regi pack` produces byte-deterministic PAX tarball.
- [ ] `regi pack` does NOT inject `index.json`.
