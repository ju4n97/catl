---
title: catl serve
description: Start a local read-only HTTP server for a catalog or registry
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl serve

Start a local read-only HTTP server for a catalog or registry

Starts an in-memory HTTP sidecar server on localhost for non-Go runtimes.
Auto-detects whether [dir] is a single catalog (contains catl.toml) or a multi-catalog registry.
All error responses use RFC 9457 Problem Details (application/problem+json).

Endpoints:
GET /v1/catalog # Get catalog identity, namespace, and tree hash
POST /v1/catalog/reload # Atomically reload catalog from disk
GET /v1/collections # List all collections with entity counts
GET /v1/collections/{collection} # Get collection details
GET /v1/collections/{collection}/entities # List entities in a collection
GET /v1/collections/{collection}/entities/{id} # Get entity manifest (JSON or YAML)
GET /v1/collections/{collection}/entities/{id}/files/{path...} # Stream bundle companion file
GET /v1/entities # List all entities across collections (?collection=...)
GET /v1/catalogs # List all tenant catalogs (registry mode)
POST /v1/catalogs/reload # Atomically reload all catalogs (registry mode)
ANY /v1/catalogs/{catalog}/* # Scoped access to tenant catalog routes

**Category:** Catalog

## Usage

```bash
catl serve [flags] [dir]
```

## Arguments

| Argument | Type     | Required |
| :------- | :------- | :------- |
| `[dir]`  | `string` | No       |

## Flags

| Flag                 | Description                      |
| :------------------- | :------------------------------- |
| `-a, --addr <value>` | TCP network address to listen on |

## Examples

```bash
catl serve ./flags-catalog
catl serve ./registry.d --addr 127.0.0.1:9090
```
