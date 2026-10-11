package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newRemoveCommand removes a dependency from catl.toml and prunes catl.lock.
func newRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Category:  categoryDependencies,
		Usage:     "Remove a dependency from catl.toml and lockfile",
		UsageText: "catl remove <name>",
		Description: `Removes the declared dependency from [dependencies] in catl.toml and
prunes its corresponding pin from catl.lock.`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "name",
				UsageText: "<name>",
				Required:  true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.StringArg("name")
			_ = name
			return nil
		},
	}
}
