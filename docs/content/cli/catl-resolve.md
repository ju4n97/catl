---
title: catl resolve
description: Print the filesystem path of a manifest
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl resolve

Print the filesystem path of a manifest

Returns the absolute filesystem path to an entity's manifest.
Rejects bare identifiers with a suggestion for the expected qualified ID.

**Category:** Entities

## Usage

```bash
catl resolve <collection> <qualified-id>
```

## Arguments

| Argument         | Type     | Required |
| :--------------- | :------- | :------- |
| `<collection>`   | `string` | Yes      |
| `<qualified-id>` | `string` | Yes      |

## Examples

```bash
catl resolve services platform-services.auth
$EDITOR "$(catl resolve rules platform-rules.require-auth)"
```
