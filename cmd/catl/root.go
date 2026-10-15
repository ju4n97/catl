package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl/internal/version"
)

// newRootCommand initializes the top-level catl CLI application.
func newRootCommand() *cli.Command {
	cli.VersionPrinter = func(cmd *cli.Command) {
		_, _ = fmt.Fprintln(cmd.Root().Writer, version.String())
	}

	return &cli.Command{
		Name:                   "catl",
		Usage:                  "Package manager and validator for declarative entity catalogs",
		UsageText:              "catl [flags] <command> [args...]",
		Version:                version.Version(),
		Suggest:                true,
		EnableShellCompletion:  true,
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "cwd",
				Aliases: []string{"C"},
				Usage:   "Change working directory before executing",
				Config:  cli.StringConfig{TrimSpace: true},
			},
			&cli.BoolFlag{
				Name:  "offline",
				Usage: "Disable all network calls and fail if dependencies are not cached",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress informational output and emit only errors",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Print detailed diagnostics and progress to stderr",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if dir := cmd.String("cwd"); dir != "" {
				if err := os.Chdir(dir); err != nil {
					return ctx, fmt.Errorf("change directory to %q: %w", dir, err)
				}
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			newVersionCommand(),

			// Catalog lifecycle
			newInitCommand(),
			newCheckCommand(),
			newCleanCommand(),
			newServeCommand(),

			// Entity queries
			newListCommand(),
			newResolveCommand(),
			newShowCommand(),

			// Dependencies
			newAddCommand(),
			newRemoveCommand(),
			newUpdateCommand(),
			newVendorCommand(),

			// Distribution
			newPackCommand(),
		},
	}
}
