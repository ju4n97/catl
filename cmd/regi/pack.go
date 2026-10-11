package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newPackCommand serializes a catalog into a reproducible PAX tar archive.
func newPackCommand() *cli.Command {
	return &cli.Command{
		Name:      "pack",
		Category:  categoryDistribution,
		Usage:     "Build a deterministic PAX tar archive of the catalog",
		UsageText: "regi pack [flags]",
		Description: `Validates the catalog and serializes it into an uncompressed, deterministic
POSIX tar archive with PAX headers. Excludes dotfiles and regi.lock.
Without -o, streams bytes to stdout (refuses if stdout is a TTY).

Examples:
  $ regi pack -o catalog-v1.0.tar
  $ regi pack | gzip > catalog-v1.0.tar.gz`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Write tar archive to file instead of stdout",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}
