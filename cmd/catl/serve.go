package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl"
	"github.com/ju4n97/catl/catlhttp"
)

// newServeCommand starts the local HTTP daemon for single catalogs or registries.
func newServeCommand() *cli.Command {
	return &cli.Command{
		Name:      "serve",
		Category:  categoryCatalog,
		Usage:     "Start a local read-only HTTP server for a catalog or registry",
		UsageText: "catl serve [flags] [dir]",
		Description: fmt.Sprintf(`Starts an in-memory HTTP sidecar server on localhost for non-Go runtimes.
Auto-detects whether [dir] is a single catalog (contains catl.toml) or a multi-catalog registry.
All error responses use RFC 9457 Problem Details (application/problem+json).

RESTful Endpoints:
%s

Examples:
  $ catl serve ./compliance-catalog
  $ catl serve ./registry.d --addr 127.0.0.1:9090`, catlhttp.EndpointsSummary()),
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "dir",
				UsageText: "[dir]",
				Required:  false,
				Config:    cli.StringConfig{TrimSpace: true},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Aliases: []string{"a"},
				Value:   "127.0.0.1:8080",
				Usage:   "TCP network address to listen on",
				Config:  cli.StringConfig{TrimSpace: true},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			dir := cmd.StringArg("dir")
			if dir == "" {
				dir = "."
			}

			addr := cmd.String("addr")

			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			// Detect single-catalog mode
			if _, err := os.Stat(filepath.Join(dir, "catl.toml")); err == nil {
				cat, err := catl.Open(ctx, dir)
				if err != nil {
					return cli.Exit(fmt.Sprintf("failed to load catalog: %v", err), 1)
				}

				if !cmd.Bool("quiet") {
					fmt.Printf("catl: serving catalog %q on http://%s\n", cat.Name(), addr)
				}

				srv := catlhttp.New(cat)
				if err := srv.Serve(ctx, addr); err != nil {
					return cli.Exit(err.Error(), 1)
				}
				return nil
			}

			// Fall back to multi-catalog registry mode
			reg, err := catl.OpenRegistry(ctx, dir)
			if err != nil {
				return cli.Exit(fmt.Sprintf("failed to load registry: %v", err), 1)
			}
			if len(reg.Names()) == 0 {
				return cli.Exit(fmt.Sprintf("no catalogs found in directory %q", dir), 1)
			}

			if !cmd.Bool("quiet") {
				fmt.Printf("catl: serving registry (%d catalogs: %s) on http://%s\n", len(reg.Names()), reg.Names(), addr)
			}

			srv := catlhttp.NewRegistry(reg)
			if err := srv.Serve(ctx, addr); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			return nil
		},
	}
}
