package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/core"
)

func GetPreparedScript(fpath, requestName string, overrides core.EnvMapValue) (string, error) {
	sqFile, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return "", err
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return "", err
	}
	script, _, err := sqFile.PrepareScript(conf, requestName, overrides)
	if err != nil {
		return "", fmt.Errorf("error occurred during script preparation: %v", err)
	}
	return script, nil
}
