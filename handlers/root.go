package handlers

import (
	"fmt"
	"os"
	"runtime"

	"github.com/EvWilson/sqump/handlers/cmder"
)

func PrintUsage() {
	fmt.Print(`Usage:
edit <squmpfile path> <request> - opens given request in your $EDITOR, saved when editor exits

env - handle operations related to the environment variables to be used in requests
|-> env set <squmpfile path> <env> <key> <value> - set the given environment mapping for the given environment
|-> env edit <target> - target may be either "core" or a squmpfile path

exec <squmpfile path> <request> - execute a given request

help - print this help diagnostic

info - print information of sqump resources
|-> info core - print core configuration information
|-> info file <squmpfile path> - print information about the given squmpfile

init - create a new default Squmpfile in the cwd

register <filename> - registers a squmpfile to be used by the application

serve - open the web view for collection editing and requests
`)
}

func BuildOps() *cmder.Op {
	rootOp := cmder.NewOp(
		"sqump",
		"short",
		"long",
		func() {},
	)

	return rootOp
}

func AssertArgLen(expectedLen int, errFuncs ...func()) {
	_, file, line, _ := runtime.Caller(1)
	if len(os.Args) != expectedLen {
		fmt.Printf("error: %s:%d: expected %d arguments, received %d\n", file, line, expectedLen, len(os.Args))
		for _, f := range errFuncs {
			f()
		}
		os.Exit(1)
	}
}

func AssertMinArgLen(minLen int, errFuncs ...func()) {
	_, file, line, _ := runtime.Caller(1)
	if len(os.Args) < minLen {
		fmt.Printf("error: %s:%d: expected at least %d arguments, received %d\n", file, line, minLen, len(os.Args))
		for _, f := range errFuncs {
			f()
		}
		os.Exit(1)
	}
}
