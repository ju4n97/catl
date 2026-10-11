package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

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
