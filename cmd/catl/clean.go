package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newCleanCommand removes the local download and build cache.
func newCleanCommand() *cli.Command {
	return &cli.Command{
		Name:      "clean",
		Category:  categoryCatalog,
		Usage:     "Clear local download and build cache",
		UsageText: "catl clean",
		Description: `Empties the local user cache directory ($XDG_CACHE_HOME/catl).
Does not modify .catl/vendor/ directories inside catalogs.`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}
