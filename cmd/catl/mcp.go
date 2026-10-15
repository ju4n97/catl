package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/urfave/cli/v3"

	"github.com/ju4n97/catl"
	"github.com/ju4n97/catl/catlmcp"
)

// newMCPCommand starts an official Model Context Protocol server over stdio.
func newMCPCommand() *cli.Command {
	return &cli.Command{
		Name:      "mcp",
		Category:  categoryCatalog,
		Usage:     "Start a Model Context Protocol (MCP) server over stdio",
		UsageText: "catl mcp [flags] [dir]",
		Description: `Starts a Model Context Protocol (MCP) server over standard I/O.
Exposes catalog manifests and companion files as MCP Resources, validation and resolution as Tools.

Examples:
  $ catl mcp ./compliance-catalog
  $ catl mcp ./registry.d`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "dir",
				UsageText: "[dir]",
				Required:  false,
				Config:    cli.StringConfig{TrimSpace: true},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			dir := cmd.StringArg("dir")
			if dir == "" {
				dir = "."
			}

			ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			transport := &mcp.StdioTransport{}

			// Detect single-catalog mode
			if _, err := os.Stat(filepath.Join(dir, "catl.toml")); err == nil {
				cat, err := catl.Open(ctx, dir)
				if err != nil {
					return cli.Exit(fmt.Sprintf("failed to load catalog: %v", err), 1)
				}

				srv := catlmcp.New(cat)
				return srv.Run(ctx, transport)
			}

			// Fall back to multi-catalog registry mode
			reg, err := catl.OpenRegistry(ctx, dir)
			if err != nil {
				return cli.Exit(fmt.Sprintf("failed to load registry: %v", err), 1)
			}

			if len(reg.Names()) == 0 {
				return cli.Exit(fmt.Sprintf("no catalogs found in directory %q", dir), 1)
			}

			srv := catlmcp.NewRegistry(reg)
			return srv.Run(ctx, transport)
		},
	}
}
