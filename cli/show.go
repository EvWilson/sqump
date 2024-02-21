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

func ShowOperation() *cmder.Op {
	return cmder.NewOp(
		"show",
		"show <collection path> <request name>, or none",
		"Shows the given request script with substitutions made, or fuzzy searches for one if no args provided",
		handleShow,
	)
}

func handleShow(ctx context.Context, args []string) error {
	overrides := ctx.Value(cmder.OverrideContextKey).(map[string]string)
	conf, err := handlers.GetConfig()
	if err != nil {
		return err
	}
	switch len(args) {
	case 0:
		filepath, requestName, err := handleShowFuzzy(conf, overrides)
		if err != nil {
			return err
		}
		prepared, err := handlers.GetPreparedScript(filepath, requestName, conf.CurrentEnv, overrides)
		if err != nil {
			return err
		}
		prnt.Println("Prepared script:")
		prnt.Println(prepared)
		return nil
	case 2:
		filepath, requestName := args[0], args[1]
		prepared, err := handlers.GetPreparedScript(filepath, requestName, conf.CurrentEnv, overrides)
		if err != nil {
			return err
		}
		prnt.Println("Prepared script:")
		prnt.Println(prepared)
		return nil
	default:
		return fmt.Errorf("expected 0 or 2 args to `exec`, got: %d", len(args))
	}
}

func handleShowFuzzy(conf *data.Config, overrides data.EnvMapValue) (string, string, error) {
	options := make([]string, 0)

	for _, path := range conf.Files {
		coll, err := handlers.GetCollection(path)
		if err != nil {
			return "", "", err
		}
		for _, req := range coll.Requests {
			options = append(options, fmt.Sprintf("%s.%s", coll.Name, req.Name))
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
