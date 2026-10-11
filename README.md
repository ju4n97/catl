# regi

[![Go Reference](https://img.shields.io/badge/Go_Reference-pkg.go.dev-007D9C?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ju4n97/regi)
[![Release](https://img.shields.io/github/v/release/ju4n97/regi?style=flat-square&color=FFB000&labelColor=1A1A1A)](https://github.com/ju4n97/regi/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/ju4n97/regi/ci.yaml?style=flat-square&label=CI&labelColor=1A1A1A)](https://github.com/ju4n97/regi/actions/workflows/ci.yaml)

`regi` is a package manager for declarative data. It validates YAML and JSON manifests against JSON Schema, manages dependencies, and provides verified access to catalog contents through a CLI or Go library.

It is intended as a common substrate for building engines around declarative data. Feature flag services, AI agent runtimes, headless CMSs, and policy engines can use `regi` to manage their definitions and associated files without implementing their own validation, packaging, and dependency management.

Catalogs follow GitOps principles: definitions are maintained in version control, changes can be reviewed, and published catalogs can be pinned and reproduced independently of the runtime.

## Features

- Validates manifests, schemas, and catalog structure.
- Pins dependencies and verifies content using cryptographic hashes.
- Packages reproducible catalogs into deterministic PAX tar archives.
- Supports companion files and offline vendoring.
- Provides a standalone CLI and an embeddable Go library.
- Requires no daemon or database.

## Documentation

<https://ju4n97.github.io/regi/>

## Examples

[See the examples directory](examples/)

## License

[MIT](LICENSE)
