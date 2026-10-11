package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

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
