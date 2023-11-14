package handlers

import (
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
			func(args []string) error {
				conf, err := core.ReadConfig()
				if err != nil {
					return fmt.Errorf("error reading config:", err)
				}
				conf.PrintInfo()
				return nil
			},
		),
		cmder.NewOp(
			"file",
			"info file <squmpfile path>",
			"Print basic information about the given squmpfile",
			func(args []string) error {
				if len(args) != 1 {
					return fmt.Errorf("expected 1 argument to `info file`, got: %d", len(args))
				}
				sq, err := core.ReadSqumpfile(args[0])
				if err != nil {
					fmt.Errorf("error reading squmpfile at %s: %v\n", args[0], err)
				}
				sq.PrintInfo()
				return nil
			},
		),
	)
}

func HandleInfo(args []string) error {
	switch args[2] {
	case "core":
		conf, err := core.ReadConfig()
		if err != nil {
			fmt.Println("error reading config:", err)
			return nil
		}
		conf.PrintInfo()
	case "file":
		AssertMinArgLen(4, PrintUsage)
		sq, err := core.ReadSqumpfile(args[3])
		if err != nil {
			fmt.Printf("error reading squmpfile at %s: %v\n", args[3], err)
			return nil
		}
		sq.PrintInfo()
	}
	return nil
}
