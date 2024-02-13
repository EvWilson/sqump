package handlers

import "github.com/EvWilson/sqump/data"

func GetConfig() (*data.Config, error) {
	return data.ReadConfigFrom(data.DefaultConfigLocation())
}

func GetCollection(fpath string) (*data.Collection, error) {
	return data.ReadCollection(fpath)
}
