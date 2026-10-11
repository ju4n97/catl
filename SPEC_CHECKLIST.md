# Specification Verification Matrix

## 1. Catalog & Directory Layout

- [ ] Catalog root belongs exclusively to regi; stray non-dot files fail validation (clean checkout rule).
- [ ] Any path starting with an ASCII dot (`.`) is ignored by scanners, tree hashing, and packaging.
- [ ] Symbolic links anywhere outside dot-paths cause an immediate validation error.
- [ ] Path names with ASCII control characters (`\x00`–`\x1F`, `\x7F`) or backslashes (`\`) are rejected globally.
- [ ] All paths must be NFC-normalized UTF-8.
- [ ] Declared collection directory missing on disk is treated as a valid, empty collection (no `.gitkeep` required).
- [ ] Collection subdirectories cannot nest (entities cannot contain sub-collections).
- [ ] Flat entity is represented by a single `<id>.yaml` or `<id>.json` file and cannot have companion files.
- [ ] Bundle entity is a directory `<id>/` containing exactly one `manifest.yaml` or `manifest.json`.
- [ ] An identifier cannot coexist as both a flat file and a bundle directory in the same collection.

## 2. Configuration (`regi.toml`)

- [ ] File must be valid UTF-8 without Byte Order Mark (BOM).
- [ ] File contains either `[catalog]` or `[workspace]`, never both.
- [ ] Unknown keys at any level trigger an error indicating a newer regi version is required.
- [ ] `[catalog].name` must strictly match identifier grammar `^[a-z0-9]+(-[a-z0-9]+)*$` (max 63 bytes).
- [ ] Catalog must declare at least one collection under `[collections.<name>]`.
- [ ] Collection `schema` path is optional; if omitted, collection runs in permissive mode.
- [ ] Collection `dir` defaults to `<name>` if omitted.
- [ ] Paths in `schema` and `dir` cannot be absolute, cannot contain `..` or `.`, cannot have trailing slashes, and cannot contain dot-prefixed components.
- [ ] `dir = "."` is strictly forbidden.
- [ ] Overlapping or duplicate collection directory paths fail validation.

## 3. Schemas & Manifest Validation

- [ ] Missing schema in collection defaults to permissive schema `{}` (validates YAML/JSON syntax only).
- [ ] Offline execution: JSON Schema Draft 2020-12 metaschema is embedded; validation never touches the network.
- [ ] Local `$ref` targets resolve exclusively within catalog root; external URLs fail.
- [ ] Absolute `$id` declarations in schemas do not alter local `$ref` path resolution.
- [ ] Regular expressions in schemas use Go's RE2 engine (linear time; lookarounds and backreferences rejected).
- [ ] Strict YAML 1.2 core parsing: `on`/`off`/`yes`/`no` are parsed as strings, not booleans.
- [ ] Duplicate YAML and JSON keys are rejected.
- [ ] YAML merge keys (`<<`) and custom tags are rejected.
- [ ] Multi-document YAML streams (`---`) are rejected.
- [ ] `"format": "regi-file"` asserts target file exists inside the bundle entity.
- [ ] `"format": "regi-file"` declared inside a flat entity manifest causes a validation error.
- [ ] `"format": "regi-file"` in `oneOf`/`anyOf` is asserted only on the successful matching branch.
- [ ] Hard resource limits enforced: manifest size (4 MiB), schema size (4 MiB), YAML depth (64), alias expansions (1,000,000 values), path length (1,024 bytes).

## 4. Dependencies & Lockfile (`regi.lock`)

- [ ] Supported sources: `git` (tag, branch, rev, subdir), `url` (HTTPS tarball), and `path` (local).
- [ ] Dependency table key in `regi.toml` must strictly match the upstream `catalog.name` (no aliasing).
- [ ] Git sources peel annotated tags to commits; gitlinks (submodules) are rejected; LFS pointers treated as plain text.
- [ ] `regi add <source>` records `requested`, `resolved` commit/digest, and `tree` hash in `regi.lock`.
- [ ] Local `path` dependencies are trusted and never recorded in `regi.lock`.
- [ ] `regi.lock` is read-only for inspection and validation commands; missing or mismatched pins return `ErrLockStale`.
- [ ] `regi update` re-resolves dependencies and prunes unreferenced entries.
- [ ] `regi update` rejects force-moved Git tags unless `--force` is passed.
- [ ] `regi remove <name>` removes dependency from `regi.toml` and prunes `regi.lock`.
- [ ] Dependencies are strictly non-transitive.

## 5. Workspaces

- [ ] `[workspace]` root defines `members` array of relative paths; cannot contain `catalog`, `collections`, or `dependencies`.
- [ ] Member directories must be valid catalogs; member names must be unique across the workspace.
- [ ] Workspace members share a single `regi.lock` at the workspace root; members have no individual locks.
- [ ] Upward lockfile search climbs to nearest `regi.toml` and stops at `.git` or filesystem root.
- [ ] Members depend on each other via `path = "../sibling"`.
- [ ] `regi check` from workspace root validates all members.
- [ ] `regi check -m <name>` validates only the specified member using the workspace lock.
- [ ] In Go API, `Workspace.Member("name").Resolve()` scopes entity lookups to the member's namespace.
- [ ] `regi update` at workspace root prunes lock entries only when unreferenced by all members.

## 6. Packaging & Integrity (Tree Hash)

- [ ] Tree hash algorithm excludes root `regi.lock`.
- [ ] Tree hash algorithm excludes all dot-prefixed files and directories.
- [ ] Tree hash algorithm rejects paths with control characters (`\x00`–`\x1F`, `\x7F`) or backslashes (`\`).
- [ ] Tree hash reproducible with documented shell recipe (`LC_ALL=C sort`).
- [ ] `regi pack` serializes catalog to deterministic POSIX tar archive with PAX headers.
- [ ] `regi pack` archive excludes `regi.lock`, dotfiles, local caches, and never injects `index.json`.
- [ ] `regi pack` sets `uid=0`, `gid=0`, `mtime=0`, mode `0644`, and sorts files lexicographically.
- [ ] `regi vendor` copies remote dependencies into `.regi/vendor/`.
- [ ] `regi check` and `regi.Open()` verify `.regi/vendor/` tree hash on startup (`ErrTamperedVendor`).
- [ ] `regi --offline check` validates successfully using `.regi/vendor/` with network disabled.

## 7. CLI & Operational Behavior

- [ ] Exit codes: `0` (success), `1` (catalog/validation error), `2` (usage error), `3` (transient network error).
- [ ] `regi check` is completely silent on success and exits 0 (Unix rule of silence).
- [ ] Diagnostics on failure emit in deterministic order: path, line number, JSON Pointer, message.
- [ ] `regi list` outputs tab-separated columns (`COLLECTION  QUALIFIED-ID  PATH`).
- [ ] `regi list` supports `--json` (newline-delimited JSON) and `-z` (NUL-delimited records).
- [ ] `regi list` flags `--json` and `--null` are mutually exclusive.
- [ ] `regi resolve` with bare ID exits 1 with candidate qualified ID suggestion.
- [ ] `regi show` streams verified manifest or companion file to stdout.
- [ ] `regi clean` purges local cache directory, strictly preserving `.regi/vendor/`.
- [ ] Cache writes are atomic (temporary file + rename) and read-only (`0444`).
- [ ] Standard streams honored: data to stdout, diagnostics to stderr.
