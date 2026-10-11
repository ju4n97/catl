---
title: regi list
description: List entities across collections
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi list

List entities across collections

Lists entities sorted by collection and qualified identifier.
Outputs tab-separated columns: COLLECTION QUALIFIED-ID PATH

**Category:** Entities \
**Aliases:** `regi ls`

## Usage

```bash
regi list [flags] [collection]
```

## Arguments

| Argument       | Type     | Required |
| :------------- | :------- | :------- |
| `[collection]` | `string` | No       |

## Flags

| Flag         | Description                                           |
| :----------- | :---------------------------------------------------- |
| `-a, --all`  | Include entities imported from declared dependencies  |
| `--json`     | Output entities as newline-delimited JSON objects     |
| `-z, --null` | Delimit records with ASCII NUL characters for scripts |

## Examples

```bash
regi list
regi list services
regi list --all --json
```
