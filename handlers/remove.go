package handlers

import "github.com/EvWilson/sqump/core"

func RemoveRequest(fpath, requestTitle string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	return sq.RemoveRequest(requestTitle)
}
