package handlers

import (
	"fmt"
	"os"

	"github.com/EvWilson/sqump/core"
)

func AddFile(title string) error {
	sq := core.DefaultSqumpFile()
	return AddFileAtPath(title, sq.Path)
}

func AddFileAtPath(title string, path string) error {
	sq := core.DefaultSqumpFile()
	sq.Title = title
	sq.Path = path
	if _, err := os.Stat(sq.Path); !os.IsNotExist(err) {
		return fmt.Errorf("file already exists at '%s'", sq.Path)
	}
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
