//go:build docs

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"
)

const docsOutputDir = "./docs/content/reference"

func main() {
	if err := generateDocs(newRootCommand()); err != nil {
		fmt.Fprintf(os.Stderr, "error generating docs: %v\n", err)
		os.Exit(1)
	}
}

func generateDocs(root *cli.Command) error {
	if err := os.MkdirAll(docsOutputDir, 0o755); err != nil {
		return fmt.Errorf("create docs directory: %w", err)
	}

	entries, err := os.ReadDir(docsOutputDir)
	if err != nil {
		return fmt.Errorf("read docs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			if err := os.Remove(filepath.Join(docsOutputDir, entry.Name())); err != nil {
				return fmt.Errorf("remove old %s: %w", entry.Name(), err)
			}
		}
	}

	md, err := docs.ToMarkdown(root)
	if err != nil {
		return fmt.Errorf("generate Markdown: %w", err)
	}

	path := filepath.Join(docsOutputDir, "cli.md")
	if err := os.WriteFile(path, []byte(md), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}
