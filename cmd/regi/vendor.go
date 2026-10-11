package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

// newVendorCommand copies remote dependencies into .regi/vendor for offline builds.
func newVendorCommand() *cli.Command {
	return &cli.Command{
		Name:      "vendor",
		Category:  categoryDependencies,
		Usage:     "Vendor dependencies into .regi/vendor for offline use",
		UsageText: "regi vendor [flags]",
		Description: `Copies all remote dependencies and referenced blobs into .regi/vendor/.
Local path dependencies are not copied.`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-blobs",
				Usage: "Skip downloading large blob payloads when vendoring",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}
