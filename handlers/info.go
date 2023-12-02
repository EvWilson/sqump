package handlers

import (
	"context"
	"fmt"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"
)

func InfoOperation() *cmder.Op {
	return cmder.NewOp(
		"info",
		"info <'core' or squmpfile path>",
		"Print basic information about the requested resource",
		cmder.NewNoopHandler("info"),
		cmder.NewOp(
			"core",
			"info core",
			"Print basic information about the core configuration",
			func(_ context.Context, args []string) error {
				conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
				if err != nil {
					return fmt.Errorf("error reading config: %v", err)
				}
				conf.PrintInfo()
				return nil
			},
		),
		cmder.NewOp(
			"file",
			"info file <squmpfile path>",
			"Print basic information about the given squmpfile",
			func(_ context.Context, args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("expected 1 argument to `info file`, got: %d", len(args))
				}
				sq, err := core.ReadSqumpfile(args[0])
				if err != nil {
					return fmt.Errorf("error reading squmpfile at %s: %v\n", args[0], err)
				}
				sq.PrintInfo()
				return nil
			},
		),
	)
}
