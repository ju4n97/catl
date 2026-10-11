---
title: regi pack
description: Build a deterministic PAX tar archive of the catalog
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi pack

Build a deterministic PAX tar archive of the catalog

Validates the catalog and serializes it into an uncompressed, deterministic
POSIX tar archive with PAX headers. Excludes dotfiles and regi.lock.
Without -o, streams bytes to stdout (refuses if stdout is a TTY).

**Category:** Distribution

## Usage

```bash
regi pack [flags]
```

## Flags

| Flag                   | Description                                 |
| :--------------------- | :------------------------------------------ |
| `-o, --output <value>` | Write tar archive to file instead of stdout |

## Examples

```bash
regi pack -o catalog-v1.0.tar
regi pack | gzip > catalog-v1.0.tar.gz
```
