# NAME

catl - Package manager and validator for declarative entity catalogs

# SYNOPSIS

catl

```
[--cwd|-C]=[value]
[--quiet|-q]
```

**Usage**:

```
catl [flags] <command> [args...]
```

# GLOBAL OPTIONS

**--cwd, -C**="": Change working directory before executing

**--quiet, -q**: Suppress informational output and emit only errors

# COMMANDS

## version

Print catl version and build information

> catl version [flags]

**--verbose, -v**: Print extended build and runtime diagnostics

## init

Initialize a new catalog

> catl init [flags] [dir]

**--collection**="": Name of the initial collection to create (default: "items")

**--name**="": Catalog namespace (defaults to directory base name)

## check

Validate schemas, paths, and lockfile integrity

> catl check [flags]

## clean

Clear local download and build cache

> catl clean

## serve

Start a local read-only HTTP server for a catalog or registry

> catl serve [flags] [dir]

**--addr**="": TCP network address to listen on (default: "127.0.0.1:8080")

## mcp

Start a Model Context Protocol (MCP) server over stdio

> catl mcp [flags] [dir]

## list, ls

List entities across collections

> catl list [flags] [collection]

**--all, -a**: Include entities imported from declared dependencies

**--json**: Output entities as newline-delimited JSON objects

**--null, -z**: Delimit records with ASCII NUL characters for scripts

## resolve

Print the filesystem path of a manifest

> catl resolve <collection> <qualified-id>

## show

Stream a manifest or companion file to stdout

> catl show <collection> <qualified-id> [path]

## update

Update dependencies in catl.lock to latest revisions

> catl update [flags] [name...]

**--force**: Allow updating when a remote tag has moved to a new commit

## vendor

Vendor dependencies into .catl/vendor for offline use

> catl vendor [flags]
