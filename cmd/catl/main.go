//go:build !docs && !man

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ju4n97/catl/internal/schema"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	if err := newRootCommand().Run(ctx, os.Args); err != nil {
		if valErr, ok := errors.AsType[*schema.ValidationError](err); ok {
			fmt.Fprintln(os.Stderr, valErr.Error())
		} else {
			msg := err.Error()
			if exitCoder, ok := errors.AsType[cli.ExitCoder](err); ok {
				msg = exitCoder.Error()
			}
			if msg != "" {
				fmt.Fprintf(os.Stderr, "catl: %s\n", msg)
			}
		}
		os.Exit(1)
	}
}
