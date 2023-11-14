package handlers

import (
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

func handleExec(args []string) error {
	switch len(args) {
	case 0:
		return handleExecFuzzy()
	case 2:
		filepath, requestName := args[0], args[1]
		sqFile, err := core.ReadSqumpfile(filepath)
		if err != nil {
			return err
		}
		_, err = sqFile.ExecuteRequest(requestName, make(core.LoopChecker))
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("expected 0 or 2 args to `exec`, got: %d", len(args))
	}
}

func handleExecFuzzy() error {
	options := make([]string, 0)

	conf, err := core.ReadConfig()
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
	sq, err := conf.SqumpfileByTitle(pieces[0])
	if err != nil {
		return err
	}
	_, err = sq.ExecuteRequest(pieces[1], make(core.LoopChecker))
	if err != nil {
		return err
	}

	return nil
}
