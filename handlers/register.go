package handlers

import (
	"io/fs"
	"path/filepath"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/prnt"
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
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return err
	}
	for _, collection := range found {
		ok, err := conf.CheckForRegisteredFile(collection)
		if err != nil {
			return err
		}
		if !ok {
			prnt.Printf("registering '%s'\n", collection)
			err = conf.Register(collection)
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
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
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
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return err
	}
	return conf.Unregister(path)
}
