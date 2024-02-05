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

func UpdateCollectionName(fpath, newName string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	sq.Title = newName
	return sq.Flush()
}

func UpdateRequestName(fpath, oldName, newName string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	req, ok := sq.GetRequest(oldName)
	if !ok {
		return fmt.Errorf("UpdateRequestName: no request '%s' found in collection '%s'", oldName, sq.Title)
	}
	newReq := core.Request{
		Title:  newName,
		Script: req.Script,
	}
	err = sq.RemoveRequest(oldName)
	if err != nil {
		return err
	}
	sq = sq.UpsertRequest(&newReq)
	return sq.Flush()
}

func UpdateRequestScript(fpath, requestTitle string, newScript []string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	req, ok := sq.GetRequest(requestTitle)
	if !ok {
		return fmt.Errorf("UpdateRequestScript: no request '%s' found in collection '%s'", requestTitle, sq.Title)
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
