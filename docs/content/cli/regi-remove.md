---
title: regi remove
description: Remove a dependency from regi.toml and lockfile
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi remove

Remove a dependency from regi.toml and lockfile

Removes the declared dependency from [dependencies] in regi.toml and
prunes its corresponding pin from regi.lock.

**Category:** Dependencies \
**Aliases:** `regi rm`

## Usage

```bash
regi remove <name>
```

## Arguments

| Argument | Type     | Required |
| :------- | :------- | :------- |
| `<name>` | `string` | Yes      |
