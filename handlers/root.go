package handlers

import "fmt"

func PrintUsage() {
	fmt.Print(`Usage:
edit <squmpfile path> <request> - opens given request in your $EDITOR, saved when editor exits

env - handle operations related to the environment variables to be used in requests
|-> env set <squmpfile path> <env> <key> <value> - set the given environment mapping for the given environment

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
