package handlers

import (
	"fmt"
	"path/filepath"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"
)

func RegisterOperation() *cmder.Op {
	return cmder.NewOp(
		"register",
		"register <squmpfile path>",
		"Registers the given squmpfile in your config",
		func(args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected 1 argument to `register`, got: %d", len(args))
			}
			path, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
			if err != nil {
				return err
			}

			return conf.Register(path)
		},
	)
}

func UnregisterOperation() *cmder.Op {
	return cmder.NewOp(
		"unregister",
		"unregister <squmpfile path>",
		"Unregisters the given squmpfile from your config",
		func(args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("expected 1 argument to `unregister`, got: %d", len(args))
			}
			path, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
			if err != nil {
				return err
			}

			return conf.Unregister(path)
		},
	)
}
