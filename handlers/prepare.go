package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/exec"
)

func GetPreparedScript(fpath, requestName string, overrides data.EnvMapValue) (string, error) {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return "", err
	}
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return "", err
	}
	script, _, err := exec.PrepareScript(coll, requestName, conf, overrides)
	if err != nil {
		return "", fmt.Errorf("error occurred during script preparation: %v", err)
	}
	return script, nil
}
