package handlers

import (
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/core"
)

func EditSqumpfileEnv(fpath string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	return sq.EditEnv()
}

func EditRequest(fpath, requestName string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	err = sq.EditRequest(requestName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateRequestScript(fpath, requestTitle string, newScript []string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	req, ok := sq.GetRequest(requestTitle)
	if !ok {
		return fmt.Errorf("no request '%s' found in collection '%s'", requestTitle, sq.Title)
	}
	req.Script = core.ScriptFromString(strings.Join(newScript, "\n"))
	return sq.UpsertRequest(req).Flush()
}

func EditConfigEnv(fpath string) error {
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	err = conf.EditEnv()
	if err != nil {
		return err
	}
	return nil
}
