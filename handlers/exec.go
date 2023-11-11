package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/core"
)

func ExecuteRequest(filepath, requestName string) error {
	sqFile, err := core.ReadSqumpfile(filepath)
	if err != nil {
		return err
	}
	_, ok := sqFile.GetRequest(requestName)
	if !ok {
		return fmt.Errorf("no request of name %s found in Squmpfile at %s", requestName, filepath)
	}
	err = sqFile.ExecuteRequest(requestName)
	if err != nil {
		return err
	}
	return nil
}
