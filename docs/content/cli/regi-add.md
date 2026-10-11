---
title: regi add
description: Add a remote dependency to regi.toml and lock it
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi add

Add a remote dependency to regi.toml and lock it

Fetches the target source, verifies its content, calculates its tree hash,
pins the dependency in regi.lock, and appends it to regi.toml.

**Category:** Dependencies

## Usage

```bash
regi add <source>
```

## Arguments

| Argument   | Type     | Required |
| :--------- | :------- | :------- |
| `<source>` | `string` | Yes      |

## Examples

```bash
regi add 'git+https://github.com/acme/infra#tag=v2.0.0'
regi add 'oci://ghcr.io/acme/observability#tag=1.4.0'
regi add ../local-core
```
