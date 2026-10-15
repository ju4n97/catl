# catl

[![Go Reference](https://img.shields.io/badge/Go_Reference-pkg.go.dev-007D9C?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ju4n97/catl)
[![Release](https://img.shields.io/github/v/release/ju4n97/catl?style=flat-square&color=FFB000&labelColor=1A1A1A)](https://github.com/ju4n97/catl/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/ju4n97/catl/ci.yaml?style=flat-square&label=CI&labelColor=1A1A1A)](https://github.com/ju4n97/catl/actions/workflows/ci.yaml)

`catl` is a package manager and verified runtime substrate for declarative data stored in filesystem directories.

It organizes YAML and JSON manifests into collections, validates them against JSON Schema Draft 2020-12 using strict YAML 1.2 parsing, binds companion files into bundle directories, and computes deterministic SHA-256 tree hashes for content integrity.

Catalogs can be queried via a CLI, embedded in Go applications, served over a local HTTP daemon with atomic reloads, or mounted as a Model Context Protocol (MCP) server. Multiple isolated catalogs can be managed through a single registry.

Everything runs offline as a single static binary with no database, background daemon, or CGO dependencies.

[Documentation](https://ju4n97.github.io/catl/) · [Specification](https://ju4n97.github.io/catl/guide/spec/layout-and-manifests) · [Embedding in Go](https://ju4n97.github.io/catl/guide/reference/embedding) · [HTTP API](https://ju4n97.github.io/catl/guide/reference/http-api) · [Examples](./examples)

## Why catl? Building a Git-backed Headless CMS

Git-backed CMSs like TinaCMS and Decap CMS use Git repositories as their content store. The editor UI and publishing workflows sit on top, but underneath, you need a deterministic data layer.

Building that data layer from scratch requires solving several non-trivial problems:

- Enforcing required fields, types, and constraints before content reaches consumers.
- Applying strict YAML 1.2 parsing rules to avoid quirks (such as unquoted `no` or `off` evaluating to boolean `false`).
- Binding companion assets (Markdown bodies, images, runbooks) directly to structured metadata without losing relative path safety.
- Preventing path traversal attacks (`../`) and forbidden symlinks.
- Exposing the same verified content to Go services, frontend frameworks, shell scripts, and AI tools without duplicating parsing logic.
- Computing deterministic content hashes to verify that on-disk files match expected states.

`catl` handles this layer with a single, verified catalog model. You build the editing UI, publishing workflows, and previews; `catl` handles the data layer.

## Catalog structure

A catalog consists of a configuration file, collection schemas, and entity manifests:

```text
content/
├── catl.toml                 # Catalog identity and collection bindings
├── schemas/
│   └── article.json          # JSON Schema Draft 2020-12
└── articles/
    ├── hello-world.yaml      # Flat entity (ID: hello-world)
    └── getting-started/      # Bundle entity (ID: getting-started)
        ├── manifest.yaml     # Entity manifest
        ├── body.md           # Companion Markdown file
        └── diagram.png       # Companion asset
```

Entities take one of two shapes:

- **Flat entities:** Single manifest files placed directly in a collection directory. The filename without extension is the entity identifier.
- **Bundle entities:** Directories containing a `manifest.yaml` (or `manifest.json`) alongside arbitrary companion files. The directory name is the entity identifier.

Configure collections in `catl.toml`:

```toml
# content/catl.toml
[catalog]
name = "blog"

[collections.articles]
schema = "schemas/article.json"
```

## Walkthrough

### 1. Define the schema

Define the content type using standard JSON Schema Draft 2020-12:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["title", "author", "published"],
  "properties": {
    "title": { "type": "string" },
    "author": { "type": "string" },
    "published": { "type": "boolean" },
    "reading_time_minutes": {
      "type": "integer",
      "minimum": 1
    }
  },
  "additionalProperties": false
}
```

### 2. Author content

Create a bundle entity in `content/articles/getting-started/manifest.yaml`:

```yaml
title: Getting started with catl
author: Juan
published: true
reading_time_minutes: 5
```

The article body lives in `body.md` alongside the manifest. Assets such as diagrams can live in the same directory.

### 3. Validate

Validate the catalog:

```bash
catl -C ./content check
```

Validation runs offline and exits `0` with no output on success. Syntax errors, unknown fields, schema bounds violations, and path escapes fail immediately with source file locations and JSON pointers.

## Four interfaces in a unified engine

The same catalog can be consumed across four interfaces:

### 1. From the CLI

```bash
# List all entities across collections
catl -C ./content list

# Print the absolute filesystem path of a manifest
catl -C ./content resolve articles blog.getting-started

# Stream manifest content to stdout
catl -C ./content show articles blog.getting-started

# Stream a companion file to stdout (pipe into renderers, pagers, etc.)
catl -C ./content show articles blog.getting-started body.md
```

### 2. In Go applications

Import `github.com/ju4n97/catl`. Manifests unmarshal into typed structs, and bundle files are accessible via standard `io/fs.FS`:

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

  cat, err := catl.Open(ctx, "./content", catl.Options{Offline: true})
  if err != nil {
    log.Fatalf("failed to open catalog: %v", err)
  }

  // Resolve entity by collection and qualified ID ("<namespace>.<id>")
  entity, err := cat.Resolve("articles", "blog.getting-started")
  if err != nil {
    log.Fatalf("resolve error: %v", err)
  }

  var article Article
  if err := entity.Unmarshal(&article); err != nil {
    log.Fatalf("unmarshal error: %v", err)
  }
  fmt.Printf("%s by %s\n", article.Title, article.Author)

  // Access companion files through standard io/fs.FS
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

### 3. Over HTTP (`catl serve`)

Run the read-only HTTP daemon for services written in other languages:

```bash
catl serve ./content --addr 127.0.0.1:8080
```

The daemon provides RESTful endpoints, and atomic reloads:

```bash
# Get complete entity resource (metadata + parsed manifest JSON)
curl http://127.0.0.1:8080/v1/collections/articles/entities/blog.getting-started

# Stream raw manifest file as stored on disk
curl http://127.0.0.1:8080/v1/collections/articles/entities/blog.getting-started/manifest

# List companion files in the bundle
curl http://127.0.0.1:8080/v1/collections/articles/entities/blog.getting-started/files

# Stream companion file
curl http://127.0.0.1:8080/v1/collections/articles/entities/blog.getting-started/files/body.md

# Trigger atomic hot-reload from disk
curl -X POST http://127.0.0.1:8080/v1/catalog/reload
```

Go services can also mount the handler directly into an existing `http.ServeMux` via `catlhttp.New(cat)`.

### 4. Over Model Context Protocol (MCP)

`catl` exposes catalog entities and companion files directly to MCP-compatible clients (such as Claude Desktop and Cursor):

```bash
catl mcp ./content
```

Configure the MCP client:

```json
{
  "mcpServers": {
    "blog": {
      "command": "catl",
      "args": ["mcp", "/path/to/content"]
    }
  }
}
```

- **MCP Resources:** Manifests and companion files map to read-only URI resources (`catl://articles/blog.getting-started/body.md`).
- **MCP Prompts:** Bundle entities containing prompt templates can be inspected and invoked directly by models.

## Multi-catalog registries

When managing multiple independent catalogs (such as distinct customer repositories, clinics, or environments), organize them as siblings within a registry directory:

```text
registry.d/
├── tenant-a/
│   ├── catl.toml             # name = "tenant-a"
│   └── articles/
└── tenant-b/
    ├── catl.toml             # name = "tenant-b"
    └── articles/
```

### In Go

```go
// Open all catalogs within a registry directory
reg, err := catl.OpenRegistry(ctx, "./registry.d", catl.Options{Offline: true})
if err != nil {
  log.Fatal(err)
}

// Access isolated catalogs by namespace
tenantA, ok := reg.Catalog("tenant-a")
if ok {
  article, err := tenantA.Resolve("articles", "tenant-a.getting-started")
  if err == nil {
    fmt.Println("Found article:", article.QualifiedID)
  }
}
```

### Over HTTP

Run a single daemon to serve all tenants without identifier collisions:

```bash
catl serve ./registry.d --addr 127.0.0.1:8080
```

```http
GET /v1/catalogs
GET /v1/catalogs/tenant-a/collections/articles/entities/tenant-a.getting-started
GET /v1/catalogs/tenant-b/collections/articles/entities/tenant-b.getting-started
POST /v1/catalogs/reload
```

## Installation

### Prebuilt binaries

Download static binaries for Linux, macOS, and Windows from the [releases page](https://github.com/ju4n97/catl/releases).

### Install via Go

Requires Go 1.27+:

```bash
go install github.com/ju4n97/catl/cmd/catl@latest
```

## Documentation

<https://ju4n97.github.io/catl/>

## License

[MIT](LICENSE)
