package cli

import (
	"context"
	"os"

	"github.com/EvWilson/sqump/cli/cmder"
	"github.com/EvWilson/sqump/core"
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
			"Create a new default squmpfile in the current directory",
			func(_ context.Context, args []string) error {
				return core.WriteDefaultSqumpfile()
			},
		),
		WebOperation(),
	)
	return root
}
