package handlers

import (
	"fmt"
	"os"

	"github.com/EvWilson/sqump/core"
)

func HandleAllEnv(subcommand string, args []string) error {
	switch subcommand {
	case "set":
		core.AssertArgLen(7, PrintUsage)
		squmpPath, env, key, val := os.Args[3], os.Args[4], os.Args[5], os.Args[6]
		return HandleEnvSet(squmpPath, env, key, val)
	case "edit":
		core.AssertArgLen(4, PrintUsage)
		target := os.Args[3]
		switch target {
		case "core":
			conf, err := core.ReadConfig()
			if err != nil {
				return err
			}
			err = conf.EditEnv()
			if err != nil {
				return err
			}
			return nil
		default:
			sq, err := core.ReadSqumpfile(target)
			if err != nil {
				return err
			}
			err = sq.EditEnv(target)
			if err != nil {
				return err
			}
			return nil
		}
	default:
		return fmt.Errorf("unrecognized env subcommand: %s", subcommand)
	}
}

func HandleEnvSet(path, env, key, val string) error {
	squmpFile, err := core.ReadSqumpfile(path)
	if err != nil {
		return err
	}
	squmpFile.SetEnvVar(env, key, val)
	err = squmpFile.Flush(path)
	if err != nil {
		return err
	}
	return nil
}
