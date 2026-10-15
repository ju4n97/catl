package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl"
)

// newListCommand queries entities across declared collections.
func newListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   []string{"ls"},
		Category:  categoryEntities,
		Usage:     "List entities across collections",
		UsageText: "catl list [flags] [collection]",
		Description: `Lists entities sorted by collection and qualified identifier.
Outputs tab-separated columns: COLLECTION  QUALIFIED-ID  PATH

Examples:
  $ catl list
  $ catl list services
  $ catl list --json`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "collection",
				UsageText: "[collection]",
				Required:  false,
				Config:    cli.StringConfig{TrimSpace: true},
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
			useJSON := cmd.Bool("json")
			useNull := cmd.Bool("null")

			if useJSON && useNull {
				return errors.New("cannot use --json and --null together")
			}

			cat, err := catl.Open(ctx, ".")
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			coll := cmd.StringArg("collection")
			if coll != "" && !slices.Contains(cat.Collections(), coll) {
				return cli.Exit(fmt.Sprintf("collection %q not found", coll), 1)
			}

			var entities []*catl.Entity
			if coll != "" {
				entities = cat.Entities(coll)
			} else {
				entities = cat.Entities()
			}

			enc := json.NewEncoder(os.Stdout)
			for _, ent := range entities {
				if useJSON {
					if err := enc.Encode(ent); err != nil {
						return err
					}
				} else if useNull {
					fmt.Fprintf(os.Stdout, "%s\t%s\t%s\x00", ent.Collection, ent.QualifiedID, ent.ManifestPath)
				} else {
					fmt.Fprintf(os.Stdout, "%s\t%s\t%s\n", ent.Collection, ent.QualifiedID, ent.ManifestPath)
				}
			}

			return nil
		},
	}
}
