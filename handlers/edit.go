package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/core"
)

func HandleEdit(squmpFilePath, reqName string) {
	sq, err := core.ReadSqumpfile(squmpFilePath)
	if err != nil {
		fmt.Printf("error reading squmpfile at %s: %v\n", squmpFilePath, err)
		return
	}
	err = sq.EditRequest(squmpFilePath, reqName)
	if err != nil {
		fmt.Printf("error performing edit on squmpfile at '%s', request '%s': %v\n", squmpFilePath, reqName, err)
		return
	}
}
