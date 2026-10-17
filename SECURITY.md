# Security policy

Please report vulnerabilities privately through [GitHub Security Advisories](https://github.com/ju4n97/catl/security/advisories). Don't open public issues.

If GitHub Advisories is unavailable, email the address listed in the Git commit log.

## Supported versions

Only the latest release receives security fixes. Previous versions are not maintained.

## Reporting a vulnerability

Include:

- A description of the vulnerability and its impact.
- A minimal catalog or manifest reproducing the issue.
- Operating system and `catl version --verbose` output.

`catl` doesn't operate a bug bounty program.

## Security model

`catl` enforces these built-in defenses:

- **Path traversal:** Paths must be relative and normalized with forward slashes. Segments with `..`, backslashes, control characters, or symlinks are rejected.
- **Resource limits:** Manifest and schema files are capped at 4 MiB. YAML alias expansion is capped at 1,000,000 values, and nesting depth is capped at 64 levels.
- **Regex safety:** Schema pattern assertions run on Go's RE2 engine in linear time, preventing ReDoS.
- **Network isolation:** Loading and validation make zero network calls. Remote schema `$ref` targets are rejected.
- **Integrity verification:** Remote dependencies require a matching commit and SHA-256 tree hash in `catl.lock`. Modified vendor files fail with `ErrTamperedVendor`.
- **Line endings:** Structured files must use LF. CRLF is rejected to prevent hash divergence across platforms.

## Out of scope

- **Author provenance:** Tree hashes verify byte integrity, not author identity. Verify Git commit signatures separately.
- **Concurrent disk mutation:** `catl` assumes files remain unchanged during a read or reload.
- **Business logic:** `catl` validates schema rules and layout constraints, not application semantics.
