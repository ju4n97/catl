---
title: catl
description: Package manager and validator for declarative entity catalogs
---

<!-- Generated automatically by catl docs. Do not edit directly. -->

# catl

Package manager and validator for declarative entity catalogs

## Usage

```bash
catl [flags] <command> [args...]
```

## Flags

| Flag                | Description                                                       |
| :------------------ | :---------------------------------------------------------------- |
| `-C, --cwd <value>` | Change working directory before executing                         |
| `--offline`         | Disable all network calls and fail if dependencies are not cached |
| `-q, --quiet`       | Suppress informational output and emit only errors                |
| `-v, --verbose`     | Print detailed diagnostics and progress to stderr                 |

## Subcommands

| Command                        | Description                                                   |
| :----------------------------- | :------------------------------------------------------------ |
| [`version`](./catl-version.md) | Print catl version and build information                      |
| [`init`](./catl-init.md)       | Initialize a new catalog                                      |
| [`check`](./catl-check.md)     | Validate schemas, paths, and lockfile integrity               |
| [`clean`](./catl-clean.md)     | Clear local download and build cache                          |
| [`serve`](./catl-serve.md)     | Start a local read-only HTTP server for a catalog or registry |
| [`mcp`](./catl-mcp.md)         | Start a Model Context Protocol (MCP) server over stdio        |
| [`list`](./catl-list.md)       | List entities across collections                              |
| [`resolve`](./catl-resolve.md) | Print the filesystem path of a manifest                       |
| [`show`](./catl-show.md)       | Stream a manifest or companion file to stdout                 |
| [`add`](./catl-add.md)         | Add a remote dependency to catl.toml and lock it              |
| [`remove`](./catl-remove.md)   | Remove a dependency from catl.toml and lockfile               |
| [`update`](./catl-update.md)   | Update dependencies in catl.lock to latest revisions          |
| [`vendor`](./catl-vendor.md)   | Vendor dependencies into .catl/vendor for offline use         |
