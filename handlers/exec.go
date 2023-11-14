package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers/cmder"
)

func ExecOperation() *cmder.Op {
	return cmder.NewOp(
		"exec",
		"exec <squmpfile path> <request title>",
		"Executes the given request",
		handleExec,
	)
}

func handleExec(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 args to `exec`, got: %d", len(args))
	}
	filepath, requestName := args[0], args[1]
	sqFile, err := core.ReadSqumpfile(filepath)
	if err != nil {
		return err
	}
	_, err = sqFile.ExecuteRequest(requestName)
	if err != nil {
		return err
	}
	return nil
}
