package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/data"
)

func PrintCoreInfo() error {
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}
	conf.PrintInfo()
	return nil
}

func PrintFileInfo(fpath string) error {
	c, err := data.ReadCollection(fpath)
	if err != nil {
		return fmt.Errorf("error reading collection at %s: %v\n", fpath, err)
	}
	c.PrintInfo()
	return nil
}

func ListCollections() ([]string, error) {
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return nil, err
	}
	return conf.Files, nil
}
