//go:build e2e

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/urfave/cli/v3"
)

// TestMain registers subcommands and runs the test suite in-process.
func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"catl": runCLI,
	})
}

// TestScript discovers and executes all .txtar scripts in testdata/script.
func TestScript(t *testing.T) {
	t.Parallel()

	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "script"),
		Setup: func(env *testscript.Env) error {
			// Guarantee sandbox hermeticity so host files are never touched
			homeDir := filepath.Join(env.WorkDir, "home")
			cacheDir := filepath.Join(env.WorkDir, "cache")

			if err := os.MkdirAll(homeDir, 0o755); err != nil {
				return err
			}
			if err := os.MkdirAll(cacheDir, 0o755); err != nil {
				return err
			}

			env.Setenv("HOME", homeDir)
			env.Setenv("XDG_CACHE_HOME", cacheDir)
			env.Setenv("CATL_CACHE_DIR", cacheDir)
			env.Setenv("NO_COLOR", "1")
			return nil
		},
	})
}

// runCLI executes the catl root command in-process and exits with the target code.
func runCLI() {
	ctx := context.Background()
	cmd := newRootCommand()

	if err := cmd.Run(ctx, os.Args); err != nil {
		if exitCoder, ok := errors.AsType[cli.ExitCoder](err); ok {
			if msg := exitCoder.Error(); msg != "" {
				fmt.Fprintf(os.Stderr, "catl: %s\n", msg)
			}
			os.Exit(exitCoder.ExitCode())
		}

		fmt.Fprintf(os.Stderr, "catl: %v\n", err)
		os.Exit(1)
	}
	// A normal return implies exit code 0.
}
