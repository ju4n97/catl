---
title: catl mcp
description: Start a Model Context Protocol (MCP) server over stdio
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl mcp

Start a Model Context Protocol (MCP) server over stdio

Starts a Model Context Protocol (MCP) server over standard I/O.
Exposes catalog manifests and companion files as MCP Resources, validation and resolution as Tools.

**Category:** Catalog

## Usage

```bash
catl mcp [flags] [dir]
```

## Arguments

| Argument | Type     | Required |
| :------- | :------- | :------- |
| `[dir]`  | `string` | No       |

## Examples

```bash
catl mcp ./compliance-catalog
catl mcp ./registry.d
```
