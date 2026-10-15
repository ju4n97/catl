package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newVendorCommand copies remote dependencies into .catl/vendor for offline builds.
func newVendorCommand() *cli.Command {
	return &cli.Command{
		Name:      "vendor",
		Category:  categoryDependencies,
		Usage:     "Vendor dependencies into .catl/vendor for offline use",
		UsageText: "catl vendor [flags]",
		Description: `Copies all remote dependencies and referenced blobs into .catl/vendor/.
Local path dependencies are not copied.`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}
