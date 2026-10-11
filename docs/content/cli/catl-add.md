---
title: catl add
description: Add a remote dependency to catl.toml and lock it
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl add

Add a remote dependency to catl.toml and lock it

Fetches the target source, verifies its content, calculates its tree hash,
pins the dependency in catl.lock, and appends it to catl.toml.

**Category:** Dependencies

## Usage

```bash
catl add <source>
```

## Arguments

| Argument   | Type     | Required |
| :--------- | :------- | :------- |
| `<source>` | `string` | Yes      |

## Examples

```bash
catl add 'git+https://github.com/acme/infra#tag=v2.0.0'
catl add ../local-core
```
