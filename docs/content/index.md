---
pageType: home
title: catl
description: A tool for validating, managing, and reading declarative catalogs.

hero:
  name: catl
  text: Catalogs of declarative data
  tagline: A standalone CLI and Go library for parsing manifests, validating schemas, resolving dependencies, and verifying catalog contents.
  actions:
    - theme: brand
      text: Quick start
      link: /guide/quick-start
    - theme: alt
      text: Specification
      link: /guide/spec/layout-and-manifests

features:
  - title: Manifest validation
    details: Parse YAML and JSON manifests. Validate collections against JSON Schema Draft 2020-12, with strict parsing.
    icon: 🧩
    link: /guide/spec/layout-and-manifests

  - title: Dependency locking
    details: Declare dependencies in catl.toml. Record resolved revisions and SHA-256 tree hashes in catl.lock.
    icon: 🔗
    link: /guide/spec/dependencies

  - title: Content integrity
    details: Hash catalog files deterministically. Verify remote dependencies and vendored content against their recorded tree hashes.
    icon: 🔒
    link: /guide/spec/integrity

  - title: Offline operation
    details: Validate using local files and schemas. Vendor remote dependencies for offline use in CI and air-gapped environments.
    icon: 📦
    link: /guide/spec/dependencies

  - title: Catalog bundles
    details: Store a manifest alongside companion files such as templates, runbooks, and images. Access bundle contents through standard io/fs interfaces.
    icon: 🗂️
    link: /guide/spec/layout-and-manifests

  - title: CLI and Go API
    details: Use the standalone binary or embed catl in a Go application. Both use the same parsing and validation rules.
    icon: 🛠️
    link: /guide/reference/embedding

  - title: HTTP server
    details: Serve a catalog over a local HTTP API with atomic reloads, so edits on disk are reflected without restarting.
    icon: 🌐
    link: /guide/reference/http

  - title: Model Context Protocol
    details: Expose catalog manifests, companion files, and validation tools directly to development environments via MCP.
    icon: 🤖
    link: /guide/reference/mcp
---
