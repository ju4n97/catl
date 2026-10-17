# Contributing to catl

It's recommended that you read the [docs](https://ju4n97.github.io/catl/) and [spec](https://ju4n97.github.io/catl/spec/layout-and-manifests) to see how collections, bundles, and lockfiles work together before you open a PR.

## Architectural invariants

A few design rules are non-negotiable in this project:

- **Deterministic tree hashing:** Hashes must match across Linux, macOS, and Windows down to the exact byte. Paths are sorted byte-wise (`LC_ALL=C`), ignore dotfiles, ignore root `catl.lock`, normalize slashes, and forbid symlinks.
- **Offline runtime:** Loading and checking catalogs should never touch the network. All network calls are isolated to `catl update` and `catl vendor`.
- **Immutable runtime:** Once loaded, a `*catl.Catalog` is read-only. Don't add setters or mutation methods to `Catalog` or `Entity`. If state changes on disk, build a new catalog and swap pointers with `sync/atomic`.
- **Pure Go static binary:** No CGO (`CGO_ENABLED=0`), no libc dependencies. Build tags must stay `-tags "netgo,osusergo"`.
- **Strict LF line endings:** Manifests, schemas, and `catl.toml` must use LF (`\n`). If CRLF (`\r\n`) sneaks in, validation fails with `ErrCRLFForbidden`.
- **Clean root layout:** Only declared collection folders, schema folders, `catl.toml`, `catl.lock`, and dotfiles belong at the catalog root. Stray entries will fail the tree scanner.

## Repo layout

```text
catalog.go              Catalog loader, indexer, and query engine
entity.go               Entity representation, manifest unmarshaling, bundle filesystem
registry.go             Multi-tenant registry (in-memory and directory-backed)
errors.go               Exported sentinel errors

cmd/catl/               CLI entrypoint and subcommands (one file per command)
  category.go           Command categorization constants
  main.go               Runtime entrypoint (!docs && !man)
  main_docs.go          CLI Markdown doc generator (-tags docs)
  main_man.go           Man page generator (-tags man)
  testdata/script/      CLI E2E testscript archives (.txtar)

catlhttp/               REST HTTP server
catlmcp/                Model Context Protocol (MCP) server over stdio

internal/
  conf/                 catl.toml parser, path validation, Source sum type
  hash/                 Deterministic SHA-256 tree and file hashing
  lock/                 catl.lock parser, serializer, freshness verification
  manifest/             YAML 1.2 / JSON decoding, AST line mapping, Validator
  scan/                 Filesystem tree scanner (flat vs bundle entities)
  skel/                 Scaffolding templates for catl init
  source/               Dependency resolver (local path, cache, vendor)
  version/              Build-time version info injected via ldflags

schema/                 Embedded catl.schema.json and compiler
docs/                   Rspress documentation site (Bun)
man/                    Generated man pages
```

## Dev setup

- **Go:** 1.27 or newer
- **Task:** [Taskfile runner](https://taskfile.dev) (`task`)
- **golangci-lint:** Version 2.x
- **Bun:** Only if you're editing documentation under `docs/`

## Daily workflow

Use `task` recipes for all routine work:

```bash
# Build the binary to bin/catl (injects git version info)
task build

# Run unit tests
task test

# Run tests with the race detector (needs CGO_ENABLED=1)
task test-race

# Run CLI e2e tests
task test-e2e

# Run linters
task lint

# Format Go code and docs
task fmt

# Regenerate CLI docs and man pages
task gen

# Verify the production docs build
task build-docs
```

## Testing standards

Keep tests fast, isolated, and parallel-safe whenever possible.

### Unit tests

Maintain a 1:1 match between implementation and test files (`foo.go` and `foo_test.go`).

Use `fstest.MapFS` to mock catalogs in memory. Don't touch the physical disk in unit tests unless you're explicitly testing host filesystem quirks.

Write table-driven tests with `t.Parallel()`, and name subtests in plain English (`t.Run("rejects crlf line endings", ...)`). Add a short GoDoc comment above top-level test functions describing what you're checking.

### CLI integration tests

CLI tests run in-process using `testscript`. You'll find them as `.txtar` files in `cmd/catl/testdata/script/`.

They run in an isolated sandbox where `$HOME`, `$XDG_CACHE_HOME`, and `$CATL_CACHE_DIR` are routed to temp folders. If you touch CLI flags, stdout formatting, or exit codes, update or add a `.txtar` script to match.

## Code style and linters

Always run `task fmt && task lint` before pushing.

### Doc comments

Every exported type, func, method, or constant needs a comment. Start with the name and end with a period. Simple getters only need one line; complex loaders or validators need a couple of paragraphs covering boundaries, errors, and concurrency guarantees.

### Imports and formatting

Lines cap at 140 columns (`golines`).

Imports are split into three blocks separated by an empty line (`gci`): stdlib, external deps, and `github.com/ju4n97/catl`.

```go
import (
    "context"
    "fmt"

    "github.com/urfave/cli/v3"

    "github.com/ju4n97/catl/internal/manifest"
)
```

### Error handling

Wrap underlying errors with `%w`:

```go
return fmt.Errorf("read manifest %q: %w", path, err)
```

Keep sentinel errors lowercase without trailing punctuation (`errors.New("item not found")`). Check them with `errors.Is` or `errors.AsType`.

In `cmd/catl/`, action handlers must return `cli.Exit(msg, 1)` for user-facing errors. Never call `os.Exit` or `log.Fatal` inside library packages.

### Modern Go idioms

- `any` instead of `interface{}`
- `for i := range n` instead of C-style index loops
- Stdlib `slices` and `maps` instead of hand-rolled loops
- `for part := range strings.SplitSeq(...)` for iterative splitting
- `omitzero` in struct tags when you want to omit zero values
- Sealed interfaces get an unexported method and the `//sumtype:decl` marker

## Generated files and docs

Don't hand-edit `docs/content/reference/cli.md` or `man/catl.1`. They're generated directly from the CLI command structs.

If you change a flag, command description, or argument:

1. Edit the command in `cmd/catl/<command>.go`.
2. Run `task gen`.
3. Run `task fmt`.
4. Commit the command source alongside the generated files.

When editing documentation under `docs/content/`:

- Keep explanations direct and avoid filler text.
- Run `task build-docs` to make sure there are no dead links or broken anchors before opening a PR.

## Sending PRs

Keep PRs focused on one logical change. If you've got a feature, a refactor, and a dependency bump, split them into separate PRs.

The release changelog is generated automatically from commit messages using GoReleaser. Use conventional commit prefixes in the imperative mood:

- `feat:` for new features
- `fix:` for bug fixes
- `perf:` for performance tweaks
- `refactor:` for code cleanups that don't change behavior
- Add `!` (like `feat!:`) if you're making a breaking change

If your PR addresses an issue, link it in the description with `Closes: #<issue-number>`.

Make sure all checks pass before opening the PR:

```bash
task fmt && task lint && task test && task test-e2e
```

## Using AI tools

AI tools can be used to explore ideas, understand code, and assist with implementation, but contributions must remain human-directed and human-owned.

- You are responsible for every line you submit. You must understand your changes and verify that they work before opening a PR. Unverified contributions from autonomous agents are not accepted.
- All issues, PR titles and descriptions, commit messages, and discussion comments must be written by you, in your own words, reflecting your own understanding. Don't use AI to draft or post public-facing project communication, and especially avoid wall-of-text AI-generated summaries.
- If you configure a local `AGENTS.md`, `CLAUDE.md`, or system prompt, its rules must be consistent with this document and the spec.
- Unreviewed code dumps and large unverified diffs will be closed.

## Workspace discipline

Don't commit your personal dev environment files:

- Editor configs (`.vscode/`, `.idea/`, `.zed/`, `.cursor/`)
- Assistant instruction files (`AGENTS.md`, `CLAUDE.md`, `.opencode/`)
- Local notes, plans, and scratch files

Add them to your user-level ignore file (`git config --global core.excludesfile ~/.config/git/ignore`) or your local clone's `.git/info/exclude` file.

## License

By contributing to `catl`, you agree that your code will be licensed under the project's [MIT License](LICENSE).
