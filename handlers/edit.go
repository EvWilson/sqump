package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/core"

	"github.com/ktr0731/go-fuzzyfinder"
)

func HandleAllEdit(subcommand, arg string) error {
	conf, err := core.ReadConfig()
	if err != nil {
		return err
	}
	switch subcommand {
	case "env":
		if arg == "core" {
			err = conf.EditEnv()
			if err != nil {
				return err
			}
		} else {
			err = conf.CheckForRegisteredFile(arg)
			if err != nil {
				return err
			}
			sq, err := core.ReadSqumpfile(arg)
			if err != nil {
				return err
			}
			err = sq.EditEnv(arg)
			if err != nil {
				return err
			}
		}
		return nil
	case "req":
		err = conf.CheckForRegisteredFile(arg)
		if err != nil {
			return err
		}
		sq, err := core.ReadSqumpfile(arg)
		if err != nil {
			return err
		}

		idx, err := fuzzyfinder.Find(
			sq.Requests,
			func(i int) string {
				return sq.Requests[i].Title
			},
			fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
				if i < 0 || i >= len(sq.Requests) {
					return ""
				}
				req := sq.Requests[i]
				return fmt.Sprintf("Title: %s\nScript:\n\n%s\n", req.Title, req.Script)
			}),
		)

		err = sq.EditEnv(sq.Requests[idx].Title)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unrecognized edit subcommand: '%s'", subcommand)
	}
}
