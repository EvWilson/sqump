package handlers

import (
	"os"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"
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
			func(args []string) error {
				return core.WriteDefaultSqumpfile()
			},
		),
		WebOperation(),
	)
	return root
}
