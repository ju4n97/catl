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

These are some examples of what can be built with `regi`. The idea is to keep definitions as files and let the engine focus on what to do with them.

### [`ai-agents`](examples/ai-agents/)

Agents defined as manifests, with prompts, tool definitions, and knowledge bases alongside them. The runtime is responsible for the inference; `regi` manages the definitions and their dependencies.

### [`dynamic-forms`](examples/dynamic-forms/)

A form engine driven by manifests instead of hardcoded UI components. Define fields, validation rules, and reusable components, then change forms through version-controlled data rather than application releases.

### [`exoplanet-survey`](examples/exoplanet-survey/)

Manage exoplanet targets, instruments, and observation plans as structured data. A more organized alternative to keeping everything in spreadsheets and scattered research files, with schemas and versioned scientific assets.

### [`feature-flags`](examples/feature-flags/)

Feature flags and rollout rules stored as declarative entities. The evaluator loads the catalog and decides which flags are enabled. Definitions can be updated and reviewed separately from the engine itself.

### [`headless-cms`](examples/headless-cms/)

Articles, pages, and media stored as catalog entities. Content lives in files and can be reviewed through Git, instead of being tied to a particular CMS database. The application still handles the API and rendering.

### [`localization`](examples/localization/)

Translation messages and locale-specific resources managed as catalogs. Useful when translations need their own versioning and shouldn't have to wait for an application release.

### [`observability`](examples/observability/)

Alert rules, dashboards, monitoring configuration, and runbooks in one place. Share definitions between services and keep them under version control; the monitoring system still does the actual evaluation.

### [`policies`](examples/policies/)

Security policies, authorization rules, and compliance controls packaged as reusable entities. Dependencies let different applications use shared policy catalogs without copying files between repositories.

## License

[MIT](LICENSE)
