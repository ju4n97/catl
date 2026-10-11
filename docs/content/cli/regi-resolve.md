---
title: regi resolve
description: Print the filesystem path of a manifest
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi resolve

Print the filesystem path of a manifest

Returns the absolute filesystem path to an entity's manifest.
Rejects bare identifiers with a suggestion for the expected qualified ID.

**Category:** Entities

## Usage

```bash
regi resolve <collection> <qualified-id>
```

## Arguments

| Argument         | Type     | Required |
| :--------------- | :------- | :------- |
| `<collection>`   | `string` | Yes      |
| `<qualified-id>` | `string` | Yes      |

## Examples

```bash
regi resolve services platform-services.auth
$EDITOR "$(regi resolve rules platform-rules.require-auth)"
```
