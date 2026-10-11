---
title: catl pack
description: Build a deterministic PAX tar archive of the catalog
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl pack

Build a deterministic PAX tar archive of the catalog

Validates the catalog and serializes it into an uncompressed, deterministic
POSIX tar archive with PAX headers. Excludes dotfiles and catl.lock.
Without -o, streams bytes to stdout (refuses if stdout is a TTY).

**Category:** Distribution

## Usage

```bash
catl pack [flags]
```

## Flags

| Flag                   | Description                                 |
| :--------------------- | :------------------------------------------ |
| `-o, --output <value>` | Write tar archive to file instead of stdout |

## Examples

```bash
catl pack -o catalog-v1.0.tar
catl pack | gzip > catalog-v1.0.tar.gz
```
