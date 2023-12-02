package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"

	"github.com/ktr0731/go-fuzzyfinder"
)

func ExecOperation() *cmder.Op {
	return cmder.NewOp(
		"exec",
		"exec <squmpfile path> <request title>, or none",
		"Executes the given request, or fuzzy searches for one if no args provided",
		handleExec,
	)
}

func handleExec(ctx context.Context, args []string) error {
	overrides := ctx.Value(cmder.OverrideContextKey).(map[string]string)
	var err error
	if err != nil {
		return err
	}
	switch len(args) {
	case 0:
		err = handleExecFuzzy(overrides)
	case 2:
		filepath, requestName := args[0], args[1]
		var sqFile *core.Squmpfile
		sqFile, err = core.ReadSqumpfile(filepath)
		if err != nil {
			return err
		}
		var conf *core.Config
		conf, err = core.ReadConfigFrom(core.DefaultConfigLocation())
		if err != nil {
			return err
		}
		_, err = sqFile.ExecuteRequest(conf, requestName, make(core.LoopChecker), overrides)
	default:
		return fmt.Errorf("expected 0 or 2 args to `exec`, got: %d", len(args))
	}
	if err != nil {
		fmt.Println("error occurred during script execution:")
		fmt.Print(err)
	}
	return nil
}

func handleExecFuzzy(overrides core.EnvMapValue) error {
	options := make([]string, 0)

	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}

	for _, path := range conf.Files {
		sq, err := core.ReadSqumpfile(path)
		if err != nil {
			return err
		}
		for _, req := range sq.Requests {
			options = append(options, fmt.Sprintf("%s.%s", sq.Title, req.Title))
		}
	}

	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string {
			return options[i]
		},
	)
	if err != nil {
		return err
	}

	option := options[idx]
	pieces := strings.Split(option, ".")
	if len(pieces) != 2 {
		return fmt.Errorf("more than a single '.' found in request identifier: '%s'", option)
	}
	squmpfileTitle, requestName := pieces[0], pieces[1]
	sq, err := conf.SqumpfileByTitle(squmpfileTitle)
	if err != nil {
		return err
	}
	_, err = sq.ExecuteRequest(conf, requestName, make(core.LoopChecker), overrides)
	if err != nil {
		return err
	}

	return nil
}
