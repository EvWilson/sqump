package cli

import (
	"context"
	"fmt"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/handlers"
)

func InfoOperation() *cmder.Op {
	return cmder.NewOp(
		"info",
		"info <'core' or collection path>",
		"Print basic information about the requested resource",
		cmder.NewNoopHandler("info"),
		cmder.NewOp(
			"core",
			"info core",
			"Print basic information about the core configuration",
			func(_ context.Context, args []string) error {
				return handlers.PrintCoreInfo()
			},
		),
		cmder.NewOp(
			"file",
			"info file <collection path>",
			"Print basic information about the given collection",
			func(_ context.Context, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("expected 1 argument to `info file`, got: %d", len(args))
				}
				return handlers.PrintFileInfo(args[0])
			},
		),
	)
}
