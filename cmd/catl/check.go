package main

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl"
	"github.com/ju4n97/catl/internal/manifest"
)

// newCheckCommand validates manifests, schemas, and dependency locks.
func newCheckCommand() *cli.Command {
	return &cli.Command{
		Name:      "check",
		Category:  categoryCatalog,
		Usage:     "Validate schemas, paths, and lockfile integrity",
		UsageText: "catl check [flags]",
		Description: `Validates all manifests against their collection schemas, checks path
invariants, verifies tree hashes, and validates vendor directories.
Exits 0 with no output on success (Unix rule of silence).`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := catl.Open(ctx, ".")
			if err != nil {
				if valErr, ok := errors.AsType[*manifest.ValidationError](err); ok {
					return cli.Exit(valErr.Error(), 1)
				}
				return cli.Exit(err.Error(), 1)
			}
			return nil
		},
	}
}
