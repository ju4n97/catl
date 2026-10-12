package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
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
			cacheDir := os.Getenv("CATL_CACHE_DIR")
			if cacheDir == "" {
				if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
					cacheDir = filepath.Join(xdg, "catl")
				} else if home, err := os.UserHomeDir(); err == nil {
					cacheDir = filepath.Join(home, ".cache", "catl")
				}
			}

			if cacheDir != "" {
				_ = os.RemoveAll(cacheDir)
			}
			return nil
		},
	}
}
