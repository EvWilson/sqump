package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/web"
	"github.com/EvWilson/sqump/web/util"
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
			core.Println("starting web server at 5309")
			go func() {
				time.Sleep(1 * time.Second)
				err := util.Open("http://localhost:5309")
				if err != nil {
					fmt.Println("error opening webview:", err)
				}
			}()
			err = http.ListenAndServe(":5309", mux)
			if err != nil {
				return err
			}
			return nil
		},
	)
}
