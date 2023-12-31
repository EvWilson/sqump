package cli

import (
	"context"
	"net/http"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/web"
)

func WebOperation() *cmder.Op {
	return cmder.NewOp(
		"webview",
		"webview",
		"Open the web UI for interacting with sqump",
		func(_ context.Context, args []string) error {
			mux, err := web.NewRouter()
			if err != nil {
				return err
			}
			err = http.ListenAndServe(":5309", mux)
			if err != nil {
				return err
			}
			return nil
		},
	)
}
