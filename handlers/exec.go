package handlers

import (
	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/exec"
)

func ExecuteRequest(fpath, requestName string, overrides data.EnvMapValue) error {
	var coll *data.Collection
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	var conf *data.Config
	conf, err = data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return err
	}
	_, err = exec.ExecuteRequest(coll, requestName, conf, overrides, exec.NewLoopChecker())
	return err
}
