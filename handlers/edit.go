package handlers

import "github.com/EvWilson/sqump/core"

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
