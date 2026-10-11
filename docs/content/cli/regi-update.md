---
title: regi update
description: Update dependencies in regi.lock to latest revisions
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi update

Update dependencies in regi.lock to latest revisions

Re-resolves remote dependencies to their latest remote revisions.
If [name...] is omitted, all dependencies are updated. Moving an existing
Git tag requires the --force flag.

**Category:** Dependencies

## Usage

```bash
regi update [flags] [name...]
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
regi update
regi update shared-infra
regi update --force shared-infra
```
