---
title: catl show
description: Stream a manifest or companion file to stdout
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl show

Stream a manifest or companion file to stdout

Streams verified content to standard output for piping into Unix utilities.
If [path] is omitted, outputs the entity manifest.

**Category:** Entities

## Usage

```bash
catl show <collection> <qualified-id> [path]
```

## Arguments

| Argument         | Type     | Required |
| :--------------- | :------- | :------- |
| `<collection>`   | `string` | Yes      |
| `<qualified-id>` | `string` | Yes      |
| `[path]`         | `string` | No       |

## Examples

```bash
catl show services platform-services.auth | jq .tier
catl show protocols clinical.emergency-guide runbook.md | less
```
