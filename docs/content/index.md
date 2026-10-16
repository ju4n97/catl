---
pageType: home
title: catl
description: Tool and Go library for validating, managing, and reading declarative catalogs.

hero:
  name: catl
  text: Catalogs of declarative data
  tagline: Standalone CLI and Go library for validating schemas, pinning dependencies, and reading directory-backed data.
  actions:
    - theme: brand
      text: Overview
      link: /overview
    - theme: alt
      text: Specification
      link: /spec/layout-and-manifests

features:
  - title: Manifest validation
    details: Parse YAML 1.2 and JSON manifests strictly, with validation against JSON Schema Draft 2020-12.
    icon: 🧩
    link: /spec/layout-and-manifests

  - title: Dependency locking
    details: Declare Git and local path dependencies in catl.toml. Pin exact commit revisions and tree hashes in catl.lock.
    icon: 🔗
    link: /spec/dependencies

  - title: Content integrity
    details: Compute deterministic SHA-256 tree hashes across platforms. Verify dependencies and vendored content before reading.
    icon: 🔒
    link: /spec/integrity

  - title: Offline operation
    details: Validate using local schemas and cache. Vendor dependencies into .catl/vendor for air-gapped environments and CI.
    icon: 📦
    link: /spec/dependencies

  - title: Catalog bundles
    details: Store manifests alongside companion files like runbooks and images. Access them through standard Go io/fs interfaces.
    icon: 🗂️
    link: /spec/layout-and-manifests

  - title: CLI and Go library
    details: Use the standalone CLI or embed catl in Go applications. Both use the same parsing engine and validation rules.
    icon: 🛠️
    link: /reference/embedding

  - title: HTTP server
    details: Serve a catalog over a local HTTP API with atomic pointer reloads, so disk changes update without downtime.
    icon: 🌐
    link: /reference/http

  - title: Model Context Protocol
    details: Expose catalog entities, companion files, and validation tools directly to development tools via stdio MCP.
    icon: 🤖
    link: /reference/mcp
---
