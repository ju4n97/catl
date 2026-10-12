package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ju4n97/catl"
	"github.com/urfave/cli/v3"
)

// newResolveCommand prints the absolute filesystem path of an entity manifest.
func newResolveCommand() *cli.Command {
	return &cli.Command{
		Name:      "resolve",
		Category:  categoryEntities,
		Usage:     "Print the filesystem path of a manifest",
		UsageText: "catl resolve <collection> <qualified-id>",
		Description: `Returns the filesystem path to an entity's manifest.
Rejects bare identifiers with a suggestion for the expected qualified ID.

Examples:
  $ catl resolve services platform-services.auth
  $ $EDITOR "$(catl resolve rules platform-rules.require-auth)"`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "collection",
				UsageText: "<collection>",
				Required:  true,
				Config:    cli.StringConfig{TrimSpace: true},
			},
			&cli.StringArg{
				Name:      "qualified-id",
				UsageText: "<qualified-id>",
				Required:  true,
				Config:    cli.StringConfig{TrimSpace: true},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			collection := cmd.StringArg("collection")
			qualifiedID := cmd.StringArg("qualified-id")

			cat, err := catl.Open(ctx, ".", catl.Options{
				Offline: cmd.Bool("offline"),
			})
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			ent, err := cat.Resolve(collection, qualifiedID)
			if err != nil {
				if errors.Is(err, catl.ErrNotQualified) {
					suggested := fmt.Sprintf("%s.%s", cat.Name(), qualifiedID)
					return cli.Exit(fmt.Sprintf("%q is not a qualified ID (want namespace.id); did you mean %q?", qualifiedID, suggested), 1)
				}
				return cli.Exit(err.Error(), 1)
			}

			fmt.Println(ent.ManifestPath())
			return nil
		},
	}
}
