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
- Packages and publishes reproducible catalogs to OCI registries.
- Supports companion files, content-addressed blobs, and offline vendoring.
- Provides a standalone CLI and an embeddable Go library.
- Requires no daemon or database.

## Documentation

<https://ju4n97.github.io/regi/>

## Examples

- [`ai-agents`](examples/ai-agents/): Agent definitions, prompts, and knowledge resources.
- [`feature-flags`](examples/feature-flags/): Feature flags and rollout rules.
- [`headless-cms`](examples/headless-cms/): Structured content and media.
- [`localization`](examples/localization/): Translations and locale-specific resources.
- [`observability`](examples/observability/): Monitoring definitions, alerts, and runbooks.
- [`policies`](examples/policies/): Security and compliance policies.

## License

[MIT](LICENSE)
