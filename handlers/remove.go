package handlers

import (
	"os"

	"github.com/EvWilson/sqump/data"
)

func RemoveRequest(fpath, requestName string) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	return coll.RemoveRequest(requestName)
}

func RemoveCollection(fpath string) error {
	err := Unregister(fpath)
	if err != nil {
		return err
	}
	return os.Remove(fpath)
}
