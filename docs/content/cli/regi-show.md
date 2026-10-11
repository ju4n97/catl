---
title: regi show
description: Stream a manifest or companion file to stdout
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi show

Stream a manifest or companion file to stdout

Streams verified content to standard output for piping into Unix utilities.
If [path] is omitted, outputs the entity manifest.
If [path] is a .blob pointer, downloads and verifies the blob before streaming.

**Category:** Entities

## Usage

```bash
regi show <collection> <qualified-id> [path]
```

## Arguments

| Argument         | Type     | Required |
| :--------------- | :------- | :------- |
| `<collection>`   | `string` | Yes      |
| `<qualified-id>` | `string` | Yes      |
| `[path]`         | `string` | No       |

## Examples

```bash
regi show services platform-services.auth | jq .tier
regi show protocols clinical.emergency-guide runbook.md | less
```
