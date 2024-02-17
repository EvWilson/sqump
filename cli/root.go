package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/data"
)

func BuildRoot() *cmder.Root {
	root := cmder.NewRoot("Welcome to sqump!", os.Stdout)
	root.Register(
		EditOperation(),
		ExecOperation(),
		AddOperation(),
		RemoveOperation(),
		AutoregisterOperation(),
		RegisterOperation(),
		UnregisterOperation(),
		ShowOperation(),
		InfoOperation(),
		cmder.NewOp(
			"init",
			"init",
			"Create a new default collection in the current directory",
			func(_ context.Context, _ []string) error {
				return data.WriteDefaultCollection()
			},
		),
		WebOperation(),
		cmder.NewOp(
			"readonly",
			"readonly",
			"Learn more about the '--readonly' option for 'webview'",
			func(_ context.Context, _ []string) error {
				fmt.Println(`Readonly mode: run 'webview' with this option to disable POST endpoints, allowing scripts to be run but not edited.`)
				return nil
			},
		),
	)
	return root
}
