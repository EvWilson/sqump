package handlers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"
)

func AutoregisterOperation() *cmder.Op {
	return cmder.NewOp(
		"autoregister",
		"autoregister",
		"Recursively search for `Squmpfile.json` to register from the current working directory",
		func(args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("expected 0 arguments to `autoregister`, got: %d", len(args))
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			found := make([]string, 0)
			err = filepath.Walk(cwd, func(path string, info fs.FileInfo, err error) error {
				if err == nil && info.Name() == "Squmpfile.json" {
					abs, err := filepath.Abs(info.Name())
					if err != nil {
						return err
					}
					found = append(found, abs)
				}
				return nil
			})
			if err != nil {
				return err
			}
			conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
			if err != nil {
				return err
			}
			for _, squmpfile := range found {
				ok, err := conf.CheckForRegisteredFile(squmpfile)
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("registering '%s'\n", squmpfile)
					err = conf.Register(squmpfile)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

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
