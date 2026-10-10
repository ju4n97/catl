package main

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"
)

const (
	categoryCatalog      = "Catalog"
	categoryEntities     = "Entities"
	categoryDependencies = "Dependencies"
	categoryDistribution = "Distribution"
	categoryStorage      = "Storage"
)

// newInitCommand scaffolds a new catalog or workspace.
func newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Category:  categoryCatalog,
		Usage:     "Initialize a new catalog or workspace",
		UsageText: "regi init [flags] [dir]",
		Description: `Creates the initial configuration and directories for a new catalog.
If --workspace is passed, initializes a multi-catalog workspace root instead.

Examples:
  $ regi init --name platform-rules --collection rules ./rules-catalog
  $ regi init --workspace ./platform-monorepo`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "dir",
				UsageText: "[dir]",
				Required:  false,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Catalog namespace (defaults to directory base name)",
			},
			&cli.StringFlag{
				Name:  "collection",
				Value: "items",
				Usage: "Name of the initial collection to create",
			},
			&cli.BoolFlag{
				Name:  "workspace",
				Usage: "Initialize a workspace root instead of a catalog",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_ = cmd.StringArg("dir")
			return nil
		},
	}
}

// newCheckCommand validates manifests, schemas, and dependency locks.
func newCheckCommand() *cli.Command {
	return &cli.Command{
		Name:      "check",
		Aliases:   []string{"c", "validate"},
		Category:  categoryCatalog,
		Usage:     "Validate schemas, paths, and lockfile integrity",
		UsageText: "regi check [flags]",
		Description: `Validates all manifests against their collection schemas, checks path
invariants, verifies tree hashes, and validates vendor directories.
Exits 0 with no output on success (Unix rule of silence).

Examples:
  $ regi check
  $ regi check --include-blobs`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "include-blobs",
				Usage: "Query remote blob stores to verify referenced blobs exist",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("include-blobs") && cmd.Root().Bool("offline") {
				return errors.New("--include-blobs cannot be used together with --offline")
			}
			return nil
		},
	}
}

// newCleanCommand removes the local download and build cache.
func newCleanCommand() *cli.Command {
	return &cli.Command{
		Name:      "clean",
		Category:  categoryCatalog,
		Usage:     "Clear local download and build cache",
		UsageText: "regi clean",
		Description: `Empties the local user cache directory ($XDG_CACHE_HOME/regi).
Does not modify .regi/vendor/ directories inside catalogs.`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}

// newListCommand queries entities across declared collections.
func newListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   []string{"ls"},
		Category:  categoryEntities,
		Usage:     "List entities across collections",
		UsageText: "regi list [flags] [collection]",
		Description: `Lists entities sorted by collection and qualified identifier.
Outputs tab-separated columns: COLLECTION  QUALIFIED-ID  PATH

Examples:
  $ regi list
  $ regi list services
  $ regi list --all --json`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "collection",
				UsageText: "[collection]",
				Required:  false,
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Include entities imported from declared dependencies",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output entities as newline-delimited JSON objects",
			},
			&cli.BoolFlag{
				Name:    "null",
				Aliases: []string{"z"},
				Usage:   "Delimit records with ASCII NUL characters for scripts",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("json") && cmd.Bool("null") {
				return errors.New("cannot use --json and --null together")
			}
			_ = cmd.StringArg("collection")
			return nil
		},
	}
}

// newResolveCommand prints the absolute filesystem path of an entity manifest.
func newResolveCommand() *cli.Command {
	return &cli.Command{
		Name:      "resolve",
		Category:  categoryEntities,
		Usage:     "Print the filesystem path of a manifest",
		UsageText: "regi resolve <collection> <qualified-id>",
		Description: `Returns the absolute filesystem path to an entity's manifest.
Rejects bare identifiers with a suggestion for the expected qualified ID.

Examples:
  $ regi resolve services platform-services.auth
  $ $EDITOR "$(regi resolve rules platform-rules.require-auth)"`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "collection",
				UsageText: "<collection>",
				Required:  true,
			},
			&cli.StringArg{
				Name:      "qualified-id",
				UsageText: "<qualified-id>",
				Required:  true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			collection := cmd.StringArg("collection")
			qualifiedID := cmd.StringArg("qualified-id")
			_ = collection
			_ = qualifiedID
			return nil
		},
	}
}

// newShowCommand streams an entity manifest or companion file to standard output.
func newShowCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Category:  categoryEntities,
		Usage:     "Stream a manifest or companion file to stdout",
		UsageText: "regi show <collection> <qualified-id> [path]",
		Description: `Streams verified content to standard output for piping into Unix utilities.
If [path] is omitted, outputs the entity manifest.
If [path] is a .blob pointer, downloads and verifies the blob before streaming.

Examples:
  $ regi show services platform-services.auth | jq .tier
  $ regi show protocols clinical.emergency-guide runbook.md | less`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "collection",
				UsageText: "<collection>",
				Required:  true,
			},
			&cli.StringArg{
				Name:      "qualified-id",
				UsageText: "<qualified-id>",
				Required:  true,
			},
			&cli.StringArg{
				Name:      "path",
				UsageText: "[path]",
				Required:  false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			collection := cmd.StringArg("collection")
			qualifiedID := cmd.StringArg("qualified-id")
			path := cmd.StringArg("path")
			_ = collection
			_ = qualifiedID
			_ = path
			return nil
		},
	}
}

// newAddCommand adds a remote dependency to regi.toml and pins it in regi.lock.
func newAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Category:  categoryDependencies,
		Usage:     "Add a remote dependency to regi.toml and lock it",
		UsageText: "regi add <source>",
		Description: `Fetches the target source, verifies its content, calculates its tree hash,
pins the dependency in regi.lock, and appends it to regi.toml.

Examples:
  $ regi add 'git+https://github.com/acme/infra#tag=v2.0.0'
  $ regi add 'oci://ghcr.io/acme/observability#tag=1.4.0'
  $ regi add ../local-core`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "source",
				UsageText: "<source>",
				Required:  true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			source := cmd.StringArg("source")
			_ = source
			return nil
		},
	}
}

// newRemoveCommand removes a dependency from regi.toml and prunes regi.lock.
func newRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Category:  categoryDependencies,
		Usage:     "Remove a dependency from regi.toml and lockfile",
		UsageText: "regi remove <name>",
		Description: `Removes the declared dependency from [dependencies] in regi.toml and
prunes its corresponding pin from regi.lock.`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "name",
				UsageText: "<name>",
				Required:  true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.StringArg("name")
			_ = name
			return nil
		},
	}
}

// newUpdateCommand re-resolves remote dependencies and rewrites regi.lock.
func newUpdateCommand() *cli.Command {
	return &cli.Command{
		Name:      "update",
		Category:  categoryDependencies,
		Usage:     "Update dependencies in regi.lock to latest revisions",
		UsageText: "regi update [flags] [name...]",
		Description: `Re-resolves remote dependencies to their latest remote revisions.
If [name...] is omitted, all dependencies are updated. Moving an existing
Git tag requires the --force flag.

Examples:
  $ regi update
  $ regi update shared-infra
  $ regi update --force shared-infra`,
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:      "name",
				UsageText: "[name...]",
				Min:       0,
				Max:       -1,
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Allow updating when a remote tag has moved to a new commit",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			names := cmd.StringArgs("name")
			_ = names
			return nil
		},
	}
}

// newVendorCommand copies remote dependencies into .regi/vendor for offline builds.
func newVendorCommand() *cli.Command {
	return &cli.Command{
		Name:      "vendor",
		Category:  categoryDependencies,
		Usage:     "Vendor dependencies into .regi/vendor for offline use",
		UsageText: "regi vendor [flags]",
		Description: `Copies all remote dependencies and referenced blobs into .regi/vendor/.
Local path dependencies are not copied.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-blobs",
				Usage: "Skip downloading large blob payloads when vendoring",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}

// newPackCommand serializes a catalog into a reproducible PAX tar archive.
func newPackCommand() *cli.Command {
	return &cli.Command{
		Name:      "pack",
		Category:  categoryDistribution,
		Usage:     "Build a deterministic PAX tar archive of the catalog",
		UsageText: "regi pack [flags]",
		Description: `Validates the catalog and serializes it into an uncompressed, deterministic
POSIX tar archive with PAX headers. Excludes dotfiles and regi.lock.
Without -o, streams bytes to stdout (refuses if stdout is a TTY).

Examples:
  $ regi pack -o catalog-v1.0.tar
  $ regi pack | gzip > catalog-v1.0.tar.gz`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Write tar archive to file instead of stdout",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}

// newPublishCommand packs and pushes a catalog as an OCI artifact.
func newPublishCommand() *cli.Command {
	return &cli.Command{
		Name:      "publish",
		Category:  categoryDistribution,
		Usage:     "Publish catalog to an OCI registry",
		UsageText: "regi publish <oci-reference>",
		Description: `Packs the catalog and pushes it to an OCI container registry.
On success, prints the immutable manifest digest to stdout.

Examples:
  $ regi publish ghcr.io/acme/platform-rules:1.4.0`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "oci-reference",
				UsageText: "<oci-reference>",
				Required:  true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			ref := cmd.StringArg("oci-reference")
			_ = ref
			return nil
		},
	}
}

// newBlobCommand namespaces content-addressed companion file utilities.
func newBlobCommand() *cli.Command {
	return &cli.Command{
		Name:      "blob",
		Category:  categoryStorage,
		Usage:     "Manage content-addressed companion files and pointer files",
		UsageText: "regi blob <command> [args...]",
		Commands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "Convert companion files into content-addressed blobs",
				UsageText: "regi blob add <file...>",
				Description: `Uploads file to [blobs].store, caches locally with read-only
permissions (0444), writes <file>.blob, and removes the original file.`,
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name:      "file",
						UsageText: "<file...>",
						Min:       1,
						Max:       -1,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					files := cmd.StringArgs("file")
					_ = files
					return nil
				},
			},
			{
				Name:      "check",
				Usage:     "Verify blob pointer syntax and check for filename collisions",
				UsageText: "regi blob check",
				Description: `Scans all .blob pointer files in bundle entities, verifies TOML syntax,
and checks that original un-pointed files don't collide on disk.`,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			},
		},
	}
}
