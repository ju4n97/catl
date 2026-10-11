package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newInitCommand scaffolds a new catalog or workspace.
func newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Category:  categoryCatalog,
		Usage:     "Initialize a new catalog or workspace",
		UsageText: "catl init [flags] [dir]",
		Description: `Creates the initial configuration and directories for a new catalog.
If --workspace is passed, initializes a multi-catalog workspace root instead.

Examples:
  $ catl init --name platform-rules --collection rules ./rules-catalog
  $ catl init --workspace ./platform-monorepo`,
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
