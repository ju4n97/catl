// Package skel scaffolds new catalogs from embedded templates, rendering
// variable placeholders in file names and file contents.
package skel

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

// Template is a supported template that can be scaffolded.
type Template string

// Supported templates.
const (
	TemplateMinimal Template = "minimal"
)

//go:embed all:templates/*
var templateFS embed.FS

// Data contains the template substitution values.
type Data struct {
	Name       string // Catalog namespace
	Collection string // Primary collection name (defaults to "items")
}

// Copy walks the specified embedded template directory, renders variable placeholders
// in file names and file contents, and streams them to destDir. It returns the list of
// created relative file paths sorted in deterministic order.
func Copy(templateName Template, destDir string, data Data) ([]string, error) {
	if data.Collection == "" {
		data.Collection = "items"
	}

	srcFS, err := fs.Sub(templateFS, path.Join("templates", string(templateName)))
	if err != nil {
		return nil, fmt.Errorf("open template %q: %w", templateName, err)
	}

	if _, err := fs.ReadDir(srcFS, "."); err != nil {
		return nil, fmt.Errorf("unknown template %q", templateName)
	}

	var createdFiles []string

	err = fs.WalkDir(srcFS, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == "." {
			return nil
		}

		destRelPath := strings.ReplaceAll(p, "__collection__", data.Collection)
		destRelPath = strings.TrimSuffix(destRelPath, ".tmpl")
		targetPath := filepath.Join(destDir, filepath.FromSlash(destRelPath))

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		srcFile, err := srcFS.Open(p)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer outFile.Close()

		if strings.HasSuffix(p, ".tmpl") {
			content, err := io.ReadAll(srcFile)
			if err != nil {
				return err
			}
			tmpl, err := template.New(p).Parse(string(content))
			if err != nil {
				return fmt.Errorf("parse template %s: %w", p, err)
			}
			if err := tmpl.Execute(outFile, data); err != nil {
				return fmt.Errorf("render template %s: %w", p, err)
			}
		} else {
			if _, err := io.Copy(outFile, srcFile); err != nil {
				return err
			}
		}

		createdFiles = append(createdFiles, filepath.ToSlash(targetPath))
		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.Sort(createdFiles)
	return createdFiles, nil
}
