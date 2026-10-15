# catl

[![Go Reference](https://img.shields.io/badge/Go_Reference-pkg.go.dev-007D9C?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ju4n97/catl)
[![Release](https://img.shields.io/github/v/release/ju4n97/catl?style=flat-square&color=FFB000&labelColor=1A1A1A)](https://github.com/ju4n97/catl/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/ju4n97/catl/ci.yaml?style=flat-square&label=CI&labelColor=1A1A1A)](https://github.com/ju4n97/catl/actions/workflows/ci.yaml)

`catl` is a package manager and runtime for structured YAML and JSON data stored in the filesystem.

It groups manifests into collections, validates them against JSON Schema Draft 2020-12 using strict YAML 1.2 parsing, and keeps companion files alongside their manifests in bundle directories. External Git dependencies are pinned in a lockfile, and deterministic SHA-256 tree hashes let you verify catalog contents.

The CLI is a single static binary. You don't need a database or background daemon, and schema validation runs entirely offline.

[Documentation](https://ju4n97.github.io/catl/) · [Layout and manifests](https://ju4n97.github.io/catl/guide/spec/layout-and-manifests) · [Embedding in Go](https://ju4n97.github.io/catl/guide/reference/embedding) · [HTTP API](https://ju4n97.github.io/catl/guide/reference/http)

## Why catl?

Git works well as a source of truth for structured content, but reading files is only part of the process. A production application still needs to validate manifests, check that referenced assets exist, resolve dependencies, and make sure different runtimes interpret the data consistently.

`catl` provides that data layer. You can build the editing UI, publishing workflow, and rendering logic yourself without having to implement the validation and filesystem rules around them.

A CMS is just one use case. The same approach works for service catalogs, feature flag registries, policy repositories, knowledge bases, AI agent registries, and many other use cases. Any system that stores structured data in Git can use the same model.

See the [quick start](https://ju4n97.github.io/catl/guide/quick-start) for a walkthrough.

## Catalog layout

A catalog starts with `catl.toml`, which declares its identity, collections, and dependencies. Collections contain entities, each represented by a manifest file or a bundle directory.

```text
content/
├── catl.toml                 # Catalog identity, collections, and dependencies
├── catl.lock                 # Pinned dependency revisions and tree hashes
├── schemas/
│   └── article.json          # JSON Schema Draft 2020-12
└── articles/
    ├── hello-world.yaml      # Flat entity: ID "hello-world"
    └── release-v1/           # Bundle entity: ID "release-v1"
        ├── manifest.yaml
        ├── body.md
        └── banner.svg
```

For example, this configuration declares an `articles` collection with a schema:

```toml
[catalog]
name = "blog"

[collections.articles]
schema = "schemas/article.json"
```

Flat entities contain only a manifest. Use a bundle when an entity needs companion files such as images, Markdown documents, PDFs, etc. See [Layout and manifests](https://ju4n97.github.io/catl/guide/spec/layout-and-manifests) for the full rules.

## Quick start

### 1. Use the CLI

Run these commands against the catalog in `./content`:

```bash
# Validate manifests and catalog layout. Silent on success.
catl -C ./content check

# List entities in deterministic order.
catl -C ./content list

# Resolve an entity to its manifest path.
catl -C ./content resolve articles blog.release-v1

# Print the manifest or a companion file.
catl -C ./content show articles blog.release-v1
catl -C ./content show articles blog.release-v1 body.md
```

The CLI can validate catalogs, resolve entities, and read manifests or companion files.

### 2. Embed it in Go

Import `github.com/ju4n97/catl` to load a catalog and work with its entities directly. Manifests can be decoded into your own structs, and bundle files are accessible through the standard `io/fs` interface.

```go
package main

import (
    "context"
    "fmt"
    "io/fs"
    "log"

    "github.com/ju4n97/catl"
)

type Article struct {
    Title     string `json:"title"`
    Author    string `json:"author"`
    Published bool   `json:"published"`
}

func main() {
    ctx := context.Background()

    cat, err := catl.Open(ctx, "./content")
    if err != nil {
        log.Fatal(err)
    }

    entity, err := cat.Resolve("articles", "blog.release-v1")
    if err != nil {
        log.Fatal(err)
    }

    var article Article
    if err := entity.Unmarshal(&article); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%s by %s\n", article.Title, article.Author)

    if entity.IsBundle() {
        bundleFS, err := entity.FS(ctx)
        if err != nil {
            log.Fatal(err)
        }

        body, err := fs.ReadFile(bundleFS, "body.md")
        if err != nil {
            log.Fatal(err)
        }

        fmt.Println(string(body))
    }
}
```

`Resolve` takes a collection name and a qualified entity ID in the form `namespace.id`. For bundle entities, `Entity.FS` gives you access to companion files without having to manage their paths yourself.

See [Embedding in Go](https://ju4n97.github.io/catl/guide/reference/embedding) for more examples.

### 3. Run the HTTP server

The HTTP server exposes a loaded catalog to other runtimes. Start it locally with:

```bash
catl serve ./content --addr 127.0.0.1:8080
```

You can then query entities and stream companion files over HTTP:

```bash
# Get the entity metadata and manifest.
curl http://127.0.0.1:8080/v1/collections/articles/entities/blog.release-v1

# Read a companion file.
curl http://127.0.0.1:8080/v1/collections/articles/entities/blog.release-v1/files/body.md

# Reload the catalog from disk.
curl -X POST http://127.0.0.1:8080/v1/catalog/reload
```

Catalog reloads use atomic replacement, so readers don't have to observe a partially loaded catalog. The [HTTP API reference](https://ju4n97.github.io/catl/guide/reference/http) documents the available endpoints.

### 4. Connect through MCP

You can also expose catalog manifests, companion files, and validation tools to MCP-compatible editors and assistants:

```bash
catl mcp ./content
```

Manifests and companion files are exposed as read-only `catl://` resources.

## Installation

<https://ju4n97.github.io/catl/guide/installation/>

## Documentation

<https://ju4n97.github.io/catl/>

## License

[MIT](LICENSE)
