---
title: catl list
description: List entities across collections
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl list

List entities across collections

Lists entities sorted by collection and qualified identifier.
Outputs tab-separated columns: COLLECTION QUALIFIED-ID PATH

**Category:** Entities \
**Aliases:** `catl ls`

## Usage

```bash
catl list [flags] [collection]
```

## Arguments

| Argument       | Type     | Required |
| :------------- | :------- | :------- |
| `[collection]` | `string` | No       |

## Flags

| Flag         | Description                                           |
| :----------- | :---------------------------------------------------- |
| `--json`     | Output entities as newline-delimited JSON objects     |
| `-z, --null` | Delimit records with ASCII NUL characters for scripts |

## Examples

```bash
catl list
catl list services
catl list --json
```
