package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl/internal/skel"
)

// newInitCommand scaffolds a new catalog.
func newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Category:  categoryCatalog,
		Usage:     "Initialize a new catalog",
		UsageText: "catl init [flags] [dir]",
		Description: `Creates the initial configuration and directories for a new catalog.

Examples:
  $ catl init --name platform-rules --collection rules ./rules-catalog`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "dir",
				UsageText: "[dir]",
				Required:  false,
				Config:    cli.StringConfig{TrimSpace: true},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "name",
				Usage:  "Catalog namespace (defaults to directory base name)",
				Config: cli.StringConfig{TrimSpace: true},
			},
			&cli.StringFlag{
				Name:   "collection",
				Value:  "items",
				Usage:  "Name of the initial collection to create",
				Config: cli.StringConfig{TrimSpace: true},
			},
			&cli.StringFlag{
				Name:    "template",
				Aliases: []string{"t"},
				Usage:   "Starter template to scaffold",
				Config:  cli.StringConfig{TrimSpace: true},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			dir := cmd.StringArg("dir")
			if dir == "" {
				dir = "."
			}

			if _, err := os.Stat(filepath.Join(dir, "catl.toml")); err == nil {
				return cli.Exit("catl.toml already exists", 1)
			}

			name := cmd.String("name")
			if name == "" {
				abs, _ := filepath.Abs(dir)
				name = filepath.Base(abs)
			}

			created, err := skel.Copy(skel.TemplateMinimal, dir, skel.Data{
				Name:       name,
				Collection: cmd.String("collection"),
			})
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			for _, path := range created {
				fmt.Printf("create %s\n", path)
			}

			return nil
		},
	}
}
