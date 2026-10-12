package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ju4n97/catl/internal/config"
	"github.com/ju4n97/catl/internal/skel"
	"github.com/urfave/cli/v3"
)

// newInitCommand scaffolds a new catalog or workspace.
func newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Category:  categoryCatalog,
		Usage:     "Initialize a new catalog or workspace",
		UsageText: "catl init [flags] [dir]",
		Description: `Creates the initial configuration and directories for a new catalog.
If --workspace is passed, initializes a multi-catalog workspace root instead.
When executed inside an existing workspace, automatically appends the new member.

Examples:
  $ catl init --name platform-rules --collection rules ./rules-catalog
  $ catl init --workspace ./platform-monorepo`,
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
				Name:   "name",
				Usage:  "Catalog namespace (defaults to directory base name)",
				Config: cli.StringConfig{TrimSpace: true},
			},
			&cli.StringFlag{
				Name:   "collection",
				Value:  "items",
				Usage:  "Name of the initial collection to create",
				Config: cli.StringConfig{TrimSpace: true},
			},
			&cli.BoolFlag{
				Name:  "workspace",
				Usage: "Initialize a workspace root instead of a catalog",
			},
			&cli.StringFlag{
				Name:    "template",
				Aliases: []string{"t"},
				Usage:   "Starter template to scaffold",
				Config:  cli.StringConfig{TrimSpace: true},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			dir := cmd.StringArg("dir")
			if dir == "" {
				dir = "."
			}

			if _, err := os.Stat(filepath.Join(dir, "catl.toml")); err == nil {
				return cli.Exit("catl.toml already exists", 1)
			}

			isWorkspace := cmd.Bool("workspace")

			// Select starter template
			tmpl := skel.TemplateMinimal
			if isWorkspace {
				tmpl = skel.TemplateWorkspaceMinimal
			}
			if t := cmd.String("template"); t != "" {
				tmpl = skel.Template(t)
			}

			name := cmd.String("name")
			if name == "" {
				abs, _ := filepath.Abs(dir)
				name = filepath.Base(abs)
			}

			created, err := skel.Copy(tmpl, dir, skel.Data{
				Name:       name,
				Collection: cmd.String("collection"),
			})
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			for _, path := range created {
				fmt.Printf("create %s\n", path)
			}

			// If creating a member catalog inside an existing workspace, append it
			if !isWorkspace && dir != "." {
				if err := appendWorkspaceMember("catl.toml", dir); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// appendWorkspaceMember detects if catl.toml in the current working directory
// is an enclosing workspace and appends the new member path while preserving comments.
func appendWorkspaceMember(wsFile, member string) error {
	data, err := os.ReadFile(wsFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // Not inside a workspace directory, nothing to update
		}
		return err
	}

	cfg, err := config.Parse(data)
	if err != nil || cfg.Workspace == nil {
		return nil // File exists but is a catalog, not a workspace
	}

	cleanMember := filepath.ToSlash(filepath.Clean(member))
	if slices.Contains(cfg.Workspace.Members, cleanMember) {
		return nil
	}

	cfg.Workspace.Members = append(cfg.Workspace.Members, cleanMember)
	slices.Sort(cfg.Workspace.Members)

	header := ""
	content := string(data)
	if before, _, ok := strings.Cut(content, "[workspace]"); ok {
		header = before
	}

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("[workspace]\nmembers = [\n")
	for _, m := range cfg.Workspace.Members {
		fmt.Fprintf(&sb, "  %q,\n", m)
	}
	sb.WriteString("]\n")

	if err := os.WriteFile(wsFile, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("update %s\n", wsFile)
	return nil
}
