# catl

[![Go Reference](https://img.shields.io/badge/Go_Reference-pkg.go.dev-007D9C?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ju4n97/catl)
[![Release](https://img.shields.io/github/v/release/ju4n97/catl?style=flat-square&color=FFB000&labelColor=1A1A1A)](https://github.com/ju4n97/catl/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/ju4n97/catl/ci.yaml?style=flat-square&label=CI&labelColor=1A1A1A)](https://github.com/ju4n97/catl/actions/workflows/ci.yaml)

`catl` is a standalone CLI and Go library that implements a self-contained, offline, deterministic data catalog engine.

It organizes YAML 1.2 and JSON manifests into collections, validates them against JSON Schema Draft 2020-12, binds companion assets into bundle directories, and pins external Git dependencies in a lockfile. Content integrity is verified across platforms using deterministic SHA-256 tree hashes.

`catl` requires no database, runs no background daemon, and makes zero network calls during validation. It is distributed as a single static binary with no CGO or libc dependencies.

[Documentation](https://ju4n97.github.io/catl/) · [Quick start](https://ju4n97.github.io/catl/guide/quick-start) · [Specification](https://ju4n97.github.io/catl/guide/spec/layout-and-manifests) · [Embedding in Go](https://ju4n97.github.io/catl/guide/reference/embedding) · [HTTP API](https://ju4n97.github.io/catl/guide/reference/http)

## Quick start

See the [Quick start guide](https://ju4n97.github.io/catl/guide/quick-start) for a walkthrough of building a Git-backed headless CMS data layer.

## Installation

Download static prebuilt binaries for Linux, macOS, and Windows from the [releases page](https://github.com/ju4n97/catl/releases), or install via Go:

```bash
go install github.com/ju4n97/catl/cmd/catl@latest
```

See [Installation](https://ju4n97.github.io/catl/guide/installation) for package managers, man pages, shell autocompletion, and container images.

## License

[MIT](LICENSE)
