---
title: catl update
description: Update dependencies in catl.lock to latest revisions
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl update

Update dependencies in catl.lock to latest revisions

Re-resolves remote dependencies to their latest remote revisions.
If [name...] is omitted, all dependencies are updated. Moving an existing
Git tag requires the --force flag.

**Category:** Dependencies

## Usage

```bash
catl update [flags] [name...]
```

## Arguments

| Argument    | Type        | Required |
| :---------- | :---------- | :------- |
| `[name...]` | `string...` | No       |

## Flags

| Flag      | Description                                                |
| :-------- | :--------------------------------------------------------- |
| `--force` | Allow updating when a remote tag has moved to a new commit |

## Examples

```bash
catl update
catl update shared-infra
catl update --force shared-infra
```
