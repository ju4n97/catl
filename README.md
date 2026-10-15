> [!IMPORTANT]
> The core catalog functionality is implemented. Dependency management, packaging, integrity checks, and cache hardening are still in progress.

# catl

[![Go Reference](https://img.shields.io/badge/Go_Reference-pkg.go.dev-007D9C?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ju4n97/catl)
[![Release](https://img.shields.io/github/v/release/ju4n97/catl?style=flat-square&color=FFB000&labelColor=1A1A1A)](https://github.com/ju4n97/catl/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/ju4n97/catl/ci.yaml?style=flat-square&label=CI&labelColor=1A1A1A)](https://github.com/ju4n97/catl/actions/workflows/ci.yaml)

`catl` is a package manager for declarative data. It's like `cargo` or `go mod`, but for directories of structured files instead of source code. It validates YAML and JSON manifests against JSON Schema, pins remote dependencies with cryptographic tree hashes, and gives you verified, offline access to catalog contents through a CLI or an embeddable Go library.

It's as a lightweight substrate for anything built on declarative data, such as AI agent runtimes, feature flag evaluators, headless CMSs, and policy engines, so you don't have to reinvent validation, locking, and integrity for every project.

## Quick start

A catalog is a directory containing a `catl.toml`, JSON schemas, and collections of manifests:

```text
catalog/
├── catl.toml                 # Name, collections, dependencies
├── schemas/
│   └── service.json          # JSON Schema Draft 2020-12
└── services/
    ├── auth.yaml             # Flat entity (ID: auth)
    └── billing/              # Bundle entity (ID: billing)
        ├── manifest.yaml
        └── runbook.md        # Companion file
```

### Validate and query from the CLI

```bash
catl check                                      # Validate (silent on success)
catl list                                       # List entities
catl resolve services platform.billing          # Print manifest path
catl show services platform.billing runbook.md  # Stream a companion file
```

### Read verified data in Go

```go
package main

import (
    "context"
    "fmt"
    "io/fs"
    "log"

    "github.com/ju4n97/catl"
)

type Service struct {
    Owner string `json:"owner"`
    Tier  int    `json:"tier"`
}

func main() {
    ctx := context.Background()

    cat, err := catl.Open(ctx, "./catalog", catl.Options{})
    if err != nil {
        log.Fatal(err)
    }

    entity, err := cat.Resolve("services", "platform.billing")
    if err != nil {
        log.Fatal(err)
    }

    var svc Service
    if err := entity.Unmarshal(&svc); err != nil {
        log.Fatal(err)
    }

    runbook, err := fs.ReadFile(entity.FS(ctx), "runbook.md")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%s (owner: %s) - %d bytes runbook\n",
    entity.QualifiedID(), svc.Owner, len(runbook))
}
```

## Documentation

<https://ju4n97.github.io/catl/>

## Examples

[See the examples directory](examples/)

## License

[MIT](LICENSE)
