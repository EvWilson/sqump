package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"
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
			func(args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("expected 1 arg in `add file`, got: %d", len(args))
				}
				title := args[0]
				sq := core.DefaultSqumpFile()
				sq.Title = title
				return sq.Flush()
			},
		),
		cmder.NewOp(
			"req",
			"add req <squmpfile path> <title>",
			"Add a new request with the given title to the given squmpfile",
			func(args []string) error {
				if len(args) != 2 {
					return fmt.Errorf("expected 2 args in `add req`, got: %d", len(args))
				}
				fpath, reqTitle := args[0], args[1]
				sq, err := core.ReadSqumpfile(fpath)
				if err != nil {
					return err
				}
				sq.Requests = append(sq.Requests, *core.NewRequest(reqTitle))
				return sq.Flush()
			},
		),
	)
}
