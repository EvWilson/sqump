package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EvWilson/sqump/data"
)

func AddFile(name string) error {
	sq := data.DefaultCollection()
	return AddFileAtPath(name, sq.Path)
}

func AddFileAtPath(name string, path string) error {
	coll := data.DefaultCollection()
	coll.Name = name
	coll.Path = path
	if _, err := os.Stat(coll.Path); !os.IsNotExist(err) {
		return fmt.Errorf("file already exists at '%s'", coll.Path)
	}
	if err := coll.Flush(); err != nil {
		return err
	}
	abs, err := filepath.Abs(coll.Path)
	if err != nil {
		return err
	}
	return Register(abs)
}

func AddRequest(fpath, requestName string) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	coll.Requests = append(coll.Requests, *data.NewRequest(requestName))
	return coll.Flush()
}
