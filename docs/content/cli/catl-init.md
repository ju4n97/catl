---
title: catl init
description: Initialize a new catalog or workspace
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl init

Initialize a new catalog or workspace

Creates the initial configuration and directories for a new catalog.
If --workspace is passed, initializes a multi-catalog workspace root instead.
When executed inside an existing workspace, automatically appends the new member.

**Category:** Catalog

## Usage

```bash
catl init [flags] [dir]
```

## Arguments

| Argument | Type     | Required |
| :------- | :------- | :------- |
| `[dir]`  | `string` | No       |

## Flags

| Flag                     | Description                                         |
| :----------------------- | :-------------------------------------------------- |
| `--name <value>`         | Catalog namespace (defaults to directory base name) |
| `--collection <value>`   | Name of the initial collection to create            |
| `--workspace`            | Initialize a workspace root instead of a catalog    |
| `-t, --template <value>` | Starter template to scaffold                        |

## Examples

```bash
catl init --name platform-rules --collection rules ./rules-catalog
catl init --workspace ./platform-monorepo
```
