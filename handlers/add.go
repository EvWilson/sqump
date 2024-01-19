package handlers

import (
	"github.com/EvWilson/sqump/core"
)

func AddFile(title string) error {
	sq := core.DefaultSqumpFile()
	sq.Title = title
	return sq.Flush()
}

func AddFileAtPath(title string, path string) error {
	sq := core.DefaultSqumpFile()
	sq.Title = title
	sq.Path = path
	return sq.Flush()
}

func AddRequest(fpath, requestName string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	sq.Requests = append(sq.Requests, *core.NewRequest(requestName))
	return sq.Flush()
}
