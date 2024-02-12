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
			"add file <collection name>",
			"Create a new collection and register it in your config",
			func(_ context.Context, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("expected 1 arg in `add file`, got: %d", len(args))
				}
				name := args[0]
				return handlers.AddFile(name)
			},
		),
		cmder.NewOp(
			"req",
			"add req <collection path> <name>",
			"Add a new request with the given name to the given collection",
			func(_ context.Context, args []string) error {
				if len(args) != 2 {
					return fmt.Errorf("expected 2 args in `add req`, got: %d", len(args))
				}
				fpath, reqName := args[0], args[1]
				return handlers.AddRequest(fpath, reqName)
			},
		),
	)
}
