package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers"
	"github.com/ktr0731/go-fuzzyfinder"
)

func ShowOperation() *cmder.Op {
	return cmder.NewOp(
		"show",
		"show <squmpfile path> <request title>, or none",
		"Shows the given request script with substitutions made, or fuzzy searches for one if no args provided",
		handleShow,
	)
}

func handleShow(ctx context.Context, args []string) error {
	overrides := ctx.Value(cmder.OverrideContextKey).(map[string]string)
	switch len(args) {
	case 0:
		filepath, requestName, err := handleShowFuzzy(overrides)
		if err != nil {
			return err
		}
		return handlers.Show(filepath, requestName, overrides)
	case 2:
		filepath, requestName := args[0], args[1]
		return handlers.Show(filepath, requestName, overrides)
	default:
		return fmt.Errorf("expected 0 or 2 args to `exec`, got: %d", len(args))
	}
}

func handleShowFuzzy(overrides core.EnvMapValue) (string, string, error) {
	options := make([]string, 0)

	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return "", "", err
	}

	for _, path := range conf.Files {
		sq, err := core.ReadSqumpfile(path)
		if err != nil {
			return "", "", err
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
		return "", "", err
	}

	option := options[idx]
	pieces := strings.Split(option, ".")
	if len(pieces) != 2 {
		return "", "", fmt.Errorf("more than a single '.' found in request identifier: '%s'", option)
	}
	return pieces[0], pieces[1], nil
}
