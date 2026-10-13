package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl/internal/version"
)

// newVersionCommand prints the application version and build metadata.
func newVersionCommand() *cli.Command {
	return &cli.Command{
		Name:      "version",
		Category:  categoryCatalog,
		Usage:     "Print catl version and build information",
		UsageText: "catl version [flags]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Print extended build and runtime diagnostics",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("verbose") {
				fmt.Println(version.Detailed())
				return nil
			}
			fmt.Println(version.String())
			return nil
		},
	}
}
