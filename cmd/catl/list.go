package main

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"
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
  $ catl list --all --json`,
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
