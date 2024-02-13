package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/handlers"

	"github.com/ktr0731/go-fuzzyfinder"
)

func EditOperation() *cmder.Op {
	return cmder.NewOp(
		"edit",
		"edit",
		"Opens selected item in your $EDITOR",
		handleGlobalEdit,
		cmder.NewOp(
			"env",
			"edit env <'core' or collection path>",
			"Opens given item's environment in your $EDITOR",
			handleEditCollectionEnv,
			cmder.NewOp(
				"core",
				"edit env core",
				"Opens core configuration environment in your $EDITOR",
				handleEditEnvCore,
			),
		),
		cmder.NewOp(
			"req",
			"edit req <collection path> <request name>",
			"Opens selected request from given collection for editing in your $EDITOR",
			handleEditReq,
		),
	)
}

func handleGlobalEdit(_ context.Context, _ []string) error {
	options := make([]string, 0)

	conf, err := handlers.GetConfig()
	if err != nil {
		return err
	}
	options = append(options, "core.env", "core.current_env")
	for _, path := range conf.Files {
		coll, err := handlers.GetCollection(path)
		if err != nil {
			return err
		}
		options = append(options, fmt.Sprintf("%s.env", coll.Name))
		options = append(options, fmt.Sprintf("%s.name", coll.Name))
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

	option := options[idx]
	if option == "core.env" {
		return conf.EditEnv()
	} else if option == "core.current_env" {
		return conf.EditCurrentEnv()
	} else if strings.HasSuffix(option, ".env") {
		name := strings.TrimSuffix(option, ".env")
		coll, err := conf.CollectionByName(name)
		if err != nil {
			return err
		}
		return coll.EditEnv()
	} else if strings.HasSuffix(option, ".name") {
		name := strings.TrimSuffix(option, ".name")
		coll, err := conf.CollectionByName(name)
		if err != nil {
			return err
		}
		return coll.EditName()
	} else {
		pieces := strings.Split(option, ".")
		if len(pieces) != 2 {
			return fmt.Errorf("more than a single '.' found in request identifier: '%s'", option)
		}
		coll, err := conf.CollectionByName(pieces[0])
		if err != nil {
			return err
		}
		return coll.EditRequest(pieces[1])
	}
}

func handleEditCollectionEnv(_ context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 arg to `edit env`, got: %d", len(args))
	}
	fpath := args[0]
	return handlers.EditCollectionEnvTUI(fpath)
}

func handleEditEnvCore(_ context.Context, _ []string) error {
	return handlers.EditConfigEnv(data.DefaultConfigLocation())
}

func handleEditReq(_ context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 args to `edit req`, got: %d", len(args))
	}
	collectionName, requestName := args[0], args[1]
	return handlers.EditRequest(collectionName, requestName)
}
