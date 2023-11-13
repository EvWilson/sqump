package handlers

import (
	"github.com/EvWilson/sqump/core"
)

func ExecuteRequest(filepath, requestName string) error {
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
