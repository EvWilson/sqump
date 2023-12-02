package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/handlers"
)

func AutoregisterOperation() *cmder.Op {
	return cmder.NewOp(
		"autoregister",
		"autoregister",
		"Recursively search for `Squmpfile.json` to register from the current working directory",
		func(_ context.Context, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("expected 0 arguments to `autoregister`, got: %d", len(args))
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			return handlers.Autoregister(cwd)
		},
	)
}

func RegisterOperation() *cmder.Op {
	return cmder.NewOp(
		"register",
		"register <squmpfile path>",
		"Registers the given squmpfile in your config",
		func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected 1 argument to `register`, got: %d", len(args))
			}
			return handlers.Register(args[0])
		},
	)
}

func UnregisterOperation() *cmder.Op {
	return cmder.NewOp(
		"unregister",
		"unregister <squmpfile path>",
		"Unregisters the given squmpfile from your config",
		func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected 1 argument to `unregister`, got: %d", len(args))
			}
			return handlers.Unregister(args[0])
		},
	)
}
