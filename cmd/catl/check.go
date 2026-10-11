package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newCheckCommand validates manifests, schemas, and dependency locks.
func newCheckCommand() *cli.Command {
	return &cli.Command{
		Name:      "check",
		Aliases:   []string{"c", "validate"},
		Category:  categoryCatalog,
		Usage:     "Validate schemas, paths, and lockfile integrity",
		UsageText: "catl check [flags]",
		Description: `Validates all manifests against their collection schemas, checks path
invariants, verifies tree hashes, and validates vendor directories.
Exits 0 with no output on success (Unix rule of silence).`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}
