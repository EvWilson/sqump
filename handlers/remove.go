package handlers

import (
	"os"

	"github.com/EvWilson/sqump/core"
)

func RemoveRequest(fpath, requestTitle string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	return sq.RemoveRequest(requestTitle)
}

func RemoveCollection(fpath string) error {
	err := Unregister(fpath)
	if err != nil {
		return err
	}
	return os.Remove(fpath)
}
