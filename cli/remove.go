package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/handlers"

	"github.com/ktr0731/go-fuzzyfinder"
)

func RemoveOperation() *cmder.Op {
	return cmder.NewOp(
		"remove",
		"remove",
		"Removes the selected resource from its parent collection",
		func(_ context.Context, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("expected 0 args to `remove`, got: %d", len(args))
			}
			return handleRemove()
		},
	)
}

func handleRemove() error {
	options := make([]string, 0)

	conf, err := handlers.GetConfig()
	if err != nil {
		return err
	}

	for _, path := range conf.Files {
		coll, err := handlers.GetCollection(path)
		if err != nil {
			return err
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
		return err
	}

	pieces := strings.Split(options[idx], ".")
	if len(pieces) != 2 {
		return fmt.Errorf("more than a single '.' found in request identifier: '%s'", options[idx])
	}
	coll, err := conf.CollectionByName(pieces[0])
	if err != nil {
		return err
	}
	return coll.RemoveRequest(pieces[1])
}
