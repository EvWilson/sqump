package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"

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
	var script string
	var err error
	if err != nil {
		return err
	}
	switch len(args) {
	case 0:
		script, err = handleShowFuzzy(overrides)
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
		script, _, err = sqFile.PrepareScript(conf, requestName, overrides)
	default:
		return fmt.Errorf("expected 0 or 2 args to `exec`, got: %d", len(args))
	}
	if err != nil {
		fmt.Println("error occurred during script execution:")
		fmt.Print(err)
	}
	fmt.Println("Prepared script:")
	fmt.Println(script)
	return nil
}

func handleShowFuzzy(overrides core.EnvMapValue) (string, error) {
	options := make([]string, 0)

	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return "", err
	}

	for _, path := range conf.Files {
		sq, err := core.ReadSqumpfile(path)
		if err != nil {
			return "", err
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
		return "", err
	}

	option := options[idx]
	pieces := strings.Split(option, ".")
	if len(pieces) != 2 {
		return "", fmt.Errorf("more than a single '.' found in request identifier: '%s'", option)
	}
	squmpfileTitle, requestName := pieces[0], pieces[1]
	sq, err := conf.SqumpfileByTitle(squmpfileTitle)
	if err != nil {
		return "", err
	}
	script, _, err := sq.PrepareScript(conf, requestName, overrides)
	if err != nil {
		return "", err
	}

	return script, nil
}
