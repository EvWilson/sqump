package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/core"
)

func PrintCoreInfo() error {
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}
	conf.PrintInfo()
	return nil
}

func PrintFileInfo(fpath string) error {
	sq, err := core.ReadSqumpfile(fpath)
	if err != nil {
		return fmt.Errorf("error reading squmpfile at %s: %v\n", fpath, err)
	}
	sq.PrintInfo()
	return nil
}
