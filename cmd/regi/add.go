package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

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
