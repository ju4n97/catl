# catl

[![Go Reference](https://img.shields.io/badge/Go_Reference-pkg.go.dev-007D9C?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ju4n97/catl)
[![Release](https://img.shields.io/github/v/release/ju4n97/catl?style=flat-square&color=FFB000&labelColor=1A1A1A)](https://github.com/ju4n97/catl/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/ju4n97/catl/ci.yaml?style=flat-square&label=CI&labelColor=1A1A1A)](https://github.com/ju4n97/catl/actions/workflows/ci.yaml)

`catl` is a package manager, validator, and reader for declarative data. It organizes YAML 1.2 and JSON manifests into collections, validates them against JSON Schema Draft 2020-12, binds companion assets into bundles, pins dependencies in a lockfile, and verifies tree integrity with deterministic SHA-256 digests.

It requires no database and no background daemon. Loading and validation run completely offline. The CLI is distributed as a single static binary with no CGO or libc dependencies.

[Documentation](https://ju4n97.github.io/catl/) · [Quick start](https://ju4n97.github.io/catl/quick-start) · [Specification](https://ju4n97.github.io/catl/spec/layout-and-manifests) · [Go library](https://ju4n97.github.io/catl/reference/go) · [CLI](https://ju4n97.github.io/catl/reference/cli) · [HTTP API](https://ju4n97.github.io/catl/reference/http) · [MCP server](https://ju4n97.github.io/catl/reference/mcp)

## Why catl

Structured data in Git works well until you need package-management mechanics: schema enforcement, companion assets, cross-repository dependencies, lockfiles, and integrity verification. `catl` provides these mechanisms without requiring a database, an external daemon, or an application-specific runtime.

Catalogs can store service definitions like policy rules, feature flags, clinical protocols, or documentation runbooks. The engine doesn't interpret what entities mean. The same catalog can be checked with the CLI, embedded into Go services, or served over HTTP and Model Context Protocol (MCP) endpoints.

## Catalog layout

A catalog is a directory containing a root `catl.toml`, schemas, and declared collection directories:

```text
catalog/
├── catl.toml                 # Catalog namespace, collections, and dependencies
├── catl.lock                 # Pinned commits and tree hashes for remote dependencies
├── schemas/
│   └── service.json          # JSON Schema Draft 2020-12
└── services/
    ├── auth.yaml             # Flat entity (ID: auth)
    └── billing/              # Bundle entity (ID: billing)
        ├── manifest.yaml     # Validated against schemas/service.json
        ├── runbook.md        # Companion file
        └── diagram.png       # Companion file
```

Flat entities are single manifest files. Bundle entities are directories containing one manifest and any number of companion files.

## Installation

Download prebuilt binaries for Linux, macOS, and Windows from the [releases page](https://github.com/ju4n97/catl/releases), or install via Go:

```bash
go install github.com/ju4n97/catl/cmd/catl@latest
```

See [Installation](https://ju4n97.github.io/catl/installation) for package managers (`.deb`, `.rpm`, `.apk`, `.pkg.tar.zst`), install scripts, container images, shell completion, and man pages.

To build a catalog from scratch, follow the [Quick start guide](https://ju4n97.github.io/catl/quick-start).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development contribution guidelines.

## License

[MIT](LICENSE)
