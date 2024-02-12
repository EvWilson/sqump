package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/prnt"

	"github.com/ktr0731/go-fuzzyfinder"
)

func ExecOperation() *cmder.Op {
	return cmder.NewOp(
		"exec",
		"exec <collection path> <request name>, or none",
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
		err = handlers.ExecuteRequest(filepath, requestName, overrides)
	default:
		return fmt.Errorf("expected 0 or 2 args to `exec`, got: %d", len(args))
	}
	if err != nil {
		prnt.Println("error occurred during script execution:")
		prnt.Println(err)
	}
	return nil
}

func handleExecFuzzy(overrides data.EnvMapValue) error {
	options := make([]string, 0)

	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return err
	}

	for _, path := range conf.Files {
		sq, err := data.ReadCollection(path)
		if err != nil {
			return err
		}
		for _, req := range sq.Requests {
			options = append(options, fmt.Sprintf("%s.%s", sq.Name, req.Name))
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
	_, requestName := pieces[0], pieces[1]
	return handlers.ExecuteRequest(conf.Files[idx], requestName, overrides)
}
