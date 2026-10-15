package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl/internal/source"
)

// newCleanCommand removes the local download and build cache.
func newCleanCommand() *cli.Command {
	return &cli.Command{
		Name:      "clean",
		Category:  categoryCatalog,
		Usage:     "Clear local download and build cache",
		UsageText: "catl clean",
		Description: `Empties the local user cache directory ($CATL_CACHE_DIR or $XDG_CACHE_HOME/catl).
Does not modify .catl/vendor/ directories inside catalogs.`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cacheDir := source.DefaultCacheDir(); cacheDir != "" {
				_ = os.RemoveAll(cacheDir)
			}
			return nil
		},
	}
}
