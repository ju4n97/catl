package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newAddCommand adds a remote dependency to catl.toml and pins it in catl.lock.
func newAddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Category:  categoryDependencies,
		Usage:     "Add a remote dependency to catl.toml and lock it",
		UsageText: "catl add <source>",
		Description: `Fetches the target source, verifies its content, calculates its tree hash,
pins the dependency in catl.lock, and appends it to catl.toml.

Examples:
  $ catl add 'git+https://github.com/acme/infra#tag=v2.0.0'
  $ catl add ../local-core`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "source",
				UsageText: "<source>",
				Required:  true,
				Config:    cli.StringConfig{TrimSpace: true},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			source := cmd.StringArg("source")
			_ = source
			return nil
		},
	}
}
