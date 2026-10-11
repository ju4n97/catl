---
title: regi
description: Package manager and validator for declarative entity catalogs
---

<!-- Generated automatically by regi docs. Do not edit directly. -->

# regi

Package manager and validator for declarative entity catalogs

## Usage

```bash
regi [flags] <command> [args...]
```

## Flags

| Flag                   | Description                                                       |
| :--------------------- | :---------------------------------------------------------------- |
| `-C, --cwd <value>`    | Change working directory before executing                         |
| `-m, --member <value>` | Target a specific workspace member (can be repeated)              |
| `--offline`            | Disable all network calls and fail if dependencies are not cached |
| `-q, --quiet`          | Suppress informational output and emit only errors                |
| `-v, --verbose`        | Print detailed diagnostics and progress to stderr                 |

## Subcommands

| Command                        | Description                                           |
| :----------------------------- | :---------------------------------------------------- |
| [`init`](./regi-init.md)       | Initialize a new catalog or workspace                 |
| [`check`](./regi-check.md)     | Validate schemas, paths, and lockfile integrity       |
| [`clean`](./regi-clean.md)     | Clear local download and build cache                  |
| [`list`](./regi-list.md)       | List entities across collections                      |
| [`resolve`](./regi-resolve.md) | Print the filesystem path of a manifest               |
| [`show`](./regi-show.md)       | Stream a manifest or companion file to stdout         |
| [`add`](./regi-add.md)         | Add a remote dependency to regi.toml and lock it      |
| [`remove`](./regi-remove.md)   | Remove a dependency from regi.toml and lockfile       |
| [`update`](./regi-update.md)   | Update dependencies in regi.lock to latest revisions  |
| [`vendor`](./regi-vendor.md)   | Vendor dependencies into .regi/vendor for offline use |
| [`pack`](./regi-pack.md)       | Build a deterministic PAX tar archive of the catalog  |
