package handlers

import (
	"fmt"
	"os"

	"github.com/EvWilson/sqump/config"
	"github.com/EvWilson/sqump/core"
)

func HandleAllEnv(subcommand string, args []string) error {
	switch subcommand {
	case "set":
		config.AssertArgLen(7, PrintUsage)
		squmpPath, env, key, val := os.Args[3], os.Args[4], os.Args[5], os.Args[6]
		return HandleEnvSet(squmpPath, env, key, val)
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
