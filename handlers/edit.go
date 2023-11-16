package handlers

import (
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"

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
			"edit env <'core' or squmpfile path>",
			"Opens given item's environment in your $EDITOR",
			handleEditSqumpfileEnv,
			cmder.NewOp(
				"core",
				"edit env core",
				"Opens core configuration environment in your $EDITOR",
				handleEditEnvCore,
			),
		),
		cmder.NewOp(
			"req",
			"edit req <squmpfile path> <request name>",
			"Opens selected request from given squmpfile for editing in your $EDITOR",
			handleEditReq,
		),
	)
}

func handleGlobalEdit(_ []string) error {
	options := make([]string, 0)

	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	options = append(options, "core.env", "core.current_env")
	for _, path := range conf.Files {
		sq, err := core.ReadSqumpfile(path)
		if err != nil {
			return err
		}
		options = append(options, fmt.Sprintf("%s.env", sq.Title))
		options = append(options, fmt.Sprintf("%s.title", sq.Title))
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
	if option == "core.env" {
		return conf.EditEnv()
	} else if option == "core.current_env" {
		return conf.EditCurrentEnv()
	} else if strings.HasSuffix(option, ".env") {
		title := strings.TrimSuffix(option, ".env")
		sq, err := conf.SqumpfileByTitle(title)
		if err != nil {
			return err
		}
		return sq.EditEnv()
	} else if strings.HasSuffix(option, ".title") {
		title := strings.TrimSuffix(option, ".title")
		sq, err := conf.SqumpfileByTitle(title)
		if err != nil {
			return err
		}
		return sq.EditTitle()
	} else {
		pieces := strings.Split(option, ".")
		if len(pieces) != 2 {
			return fmt.Errorf("more than a single '.' found in request identifier: '%s'", option)
		}
		sq, err := conf.SqumpfileByTitle(pieces[0])
		if err != nil {
			return err
		}
		return sq.EditRequest(pieces[1])
	}
}

func handleEditSqumpfileEnv(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 arg to `edit env`, got: %d", len(args))
	}
	fpath := args[0]
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	err = sq.EditEnv()
	if err != nil {
		return err
	}
	return nil
}

func handleEditEnvCore(_ []string) error {
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	err = conf.EditEnv()
	if err != nil {
		return err
	}
	return nil
}

func handleEditReq(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 args to `edit req`, got: %d", len(args))
	}
	squmpfileName, requestName := args[0], args[1]

	sq, err := core.ReadSqumpfile(squmpfileName)
	if err != nil {
		return err
	}

	err = sq.EditRequest(requestName)
	if err != nil {
		return err
	}

	return nil
}
