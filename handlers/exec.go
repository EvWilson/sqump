package handlers

import "github.com/EvWilson/sqump/core"

func ExecuteRequest(fpath, requestName string, overrides core.EnvMapValue) error {
	var sqFile *core.Squmpfile
	sqFile, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return err
	}
	var conf *core.Config
	conf, err = core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	_, err = sqFile.ExecuteRequest(conf, requestName, make(core.LoopChecker), overrides)
	return err
}
