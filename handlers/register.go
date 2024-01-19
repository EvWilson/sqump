package handlers

import (
	"io/fs"
	"path/filepath"

	"github.com/EvWilson/sqump/core"
)

func Autoregister(cwd string) error {
	found := make([]string, 0)
	err := filepath.Walk(cwd, func(path string, info fs.FileInfo, err error) error {
		if err == nil && info.Name() == "Squmpfile.json" {
			abs, err := filepath.Abs(info.Name())
			if err != nil {
				return err
			}
			found = append(found, abs)
		}
		return nil
	})
	if err != nil {
		return err
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	for _, squmpfile := range found {
		ok, err := conf.CheckForRegisteredFile(squmpfile)
		if err != nil {
			return err
		}
		if !ok {
			core.Printf("registering '%s'\n", squmpfile)
			err = conf.Register(squmpfile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Register(fpath string) error {
	path, err := filepath.Abs(fpath)
	if err != nil {
		return err
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	return conf.Register(path)
}

func Unregister(fpath string) error {
	path, err := filepath.Abs(fpath)
	if err != nil {
		return err
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	return conf.Unregister(path)
}
