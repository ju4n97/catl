package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl"
)

// newShowCommand streams an entity manifest or companion file to standard output.
func newShowCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Category:  categoryEntities,
		Usage:     "Stream a manifest or companion file to stdout",
		UsageText: "catl show <collection> <qualified-id> [path]",
		Description: `Streams verified content to standard output for piping into Unix utilities.
If [path] is omitted, outputs the entity manifest.

Examples:
  $ catl show services platform-services.auth | jq .tier
  $ catl show protocols clinical.emergency-guide runbook.md | less`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "collection",
				UsageText: "<collection>",
				Required:  true,
				Config:    cli.StringConfig{TrimSpace: true},
			},
			&cli.StringArg{
				Name:      "qualified-id",
				UsageText: "<qualified-id>",
				Required:  true,
				Config:    cli.StringConfig{TrimSpace: true},
			},
			&cli.StringArg{
				Name:      "path",
				UsageText: "[path]",
				Required:  false,
				Config:    cli.StringConfig{TrimSpace: true},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			collection := cmd.StringArg("collection")
			qualifiedID := cmd.StringArg("qualified-id")
			filePath := cmd.StringArg("path")

			cat, err := catl.Open(ctx, ".", catl.Options{
				Offline: cmd.Bool("offline"),
			})
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			ent, err := cat.Resolve(collection, qualifiedID)
			if err != nil {
				if errors.Is(err, catl.ErrNotQualified) {
					suggested := fmt.Sprintf("%s.%s", cat.Name(), qualifiedID)
					return cli.Exit(
						fmt.Sprintf("%q is not a qualified ID (want namespace.id); did you mean %q?", qualifiedID, suggested),
						1,
					)
				}
				return cli.Exit(err.Error(), 1)
			}

			// Stream raw manifest when companion path is omitted
			if filePath == "" {
				_, err := os.Stdout.Write(ent.RawManifest())
				return err
			}

			if !ent.IsBundle() {
				return cli.Exit(fmt.Sprintf("entity %q is a flat entity and has no companion files", ent.ID), 1)
			}

			bundleFS, err := ent.FS(ctx)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			file, err := bundleFS.Open(filePath)
			if err != nil {
				return cli.Exit(fmt.Sprintf("companion file %q not found in bundle: %v", filePath, err), 1)
			}
			defer file.Close()

			_, err = io.Copy(os.Stdout, file)
			return err
		},
	}
}
