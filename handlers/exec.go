package handlers

import (
	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/exec"
)

func ExecuteRequest(fpath, requestName, currentEnv string, overrides data.EnvMapValue) error {
	var coll *data.Collection
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	_, err = exec.ExecuteRequest(coll, requestName, currentEnv, overrides, exec.NewLoopChecker())
	return err
}

func CancelScripts() {
	exec.CancelScripts()
}
