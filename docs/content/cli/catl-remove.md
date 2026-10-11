---
title: catl remove
description: Remove a dependency from catl.toml and lockfile
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl remove

Remove a dependency from catl.toml and lockfile

Removes the declared dependency from [dependencies] in catl.toml and
prunes its corresponding pin from catl.lock.

**Category:** Dependencies \
**Aliases:** `catl rm`

## Usage

```bash
catl remove <name>
```

## Arguments

| Argument | Type     | Required |
| :------- | :------- | :------- |
| `<name>` | `string` | Yes      |
