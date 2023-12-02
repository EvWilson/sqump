package cli

import (
	"context"
	"fmt"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/handlers"
)

func AddOperation() *cmder.Op {
	return cmder.NewOp(
		"add",
		"add <'req' or 'file'>",
		"Add the requested resource",
		cmder.NewNoopHandler("add"),
		cmder.NewOp(
			"file",
			"add file <squmpfile title>",
			"Create a new squmpfile and register it in your config",
			func(_ context.Context, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("expected 1 arg in `add file`, got: %d", len(args))
				}
				title := args[0]
				return handlers.AddFile(title)
			},
		),
		cmder.NewOp(
			"req",
			"add req <squmpfile path> <title>",
			"Add a new request with the given title to the given squmpfile",
			func(_ context.Context, args []string) error {
				if len(args) != 2 {
					return fmt.Errorf("expected 2 args in `add req`, got: %d", len(args))
				}
				fpath, reqTitle := args[0], args[1]
				return handlers.AddRequest(fpath, reqTitle)
			},
		),
	)
}
