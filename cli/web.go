package cli

import (
	"context"
	"net/http"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/prnt"
	"github.com/EvWilson/sqump/web"
)

func WebOperation() *cmder.Op {
	return cmder.NewOp(
		"webview",
		"webview",
		"Open the web UI for interacting with sqump. Start with '--readonly' to block potentially destructive actions.",
		func(ctx context.Context, args []string) error {
			isReadonly := ctx.Value(cmder.ReadonlyContextKey).(bool)
			mux, err := web.NewRouter(isReadonly)
			if err != nil {
				return err
			}
			prnt.Println("starting web server at http://localhost:5309")
			err = http.ListenAndServe(":5309", mux)
			if err != nil {
				return err
			}
			return nil
		},
	)
}
