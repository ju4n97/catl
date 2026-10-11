package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newResolveCommand prints the absolute filesystem path of an entity manifest.
func newResolveCommand() *cli.Command {
	return &cli.Command{
		Name:      "resolve",
		Category:  categoryEntities,
		Usage:     "Print the filesystem path of a manifest",
		UsageText: "catl resolve <collection> <qualified-id>",
		Description: `Returns the absolute filesystem path to an entity's manifest.
Rejects bare identifiers with a suggestion for the expected qualified ID.

Examples:
  $ catl resolve services platform-services.auth
  $ $EDITOR "$(catl resolve rules platform-rules.require-auth)"`,
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
