---
title: catl init
description: Initialize a new catalog
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl init

Initialize a new catalog

Creates the initial configuration and directories for a new catalog.

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
| `-t, --template <value>` | Starter template to scaffold                        |

## Examples

```bash
catl init --name platform-rules --collection rules ./rules-catalog
```
