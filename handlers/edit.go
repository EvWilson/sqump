package handlers

import (
	"fmt"
	"strings"

	"github.com/EvWilson/sqump/data"
)

func EditCollectionEnvTUI(fpath string) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	return coll.EditEnv()
}

func EditRequest(fpath, requestName string) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	err = coll.EditRequest(requestName)
	if err != nil {
		return err
	}
	return nil
}

func UpdateCollectionName(fpath, newName string) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	coll.Name = newName
	return coll.Flush()
}

func UpdateCollectionEnv(fpath string, env data.EnvMap) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	coll.Environment = env
	return coll.Flush()
}

func UpdateRequestName(fpath, oldName, newName string) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	req, ok := coll.GetRequest(oldName)
	if !ok {
		return fmt.Errorf("UpdateRequestName: no request '%s' found in collection '%s'", oldName, coll.Name)
	}
	newReq := data.Request{
		Name:   newName,
		Script: req.Script,
	}
	err = coll.RemoveRequest(oldName)
	if err != nil {
		return err
	}
	coll = coll.UpsertRequest(&newReq)
	return coll.Flush()
}

func UpdateRequestScript(fpath, requestName string, newScript []string) error {
	coll, err := data.ReadCollection(fpath)
	if err != nil {
		return err
	}
	req, ok := coll.GetRequest(requestName)
	if !ok {
		return fmt.Errorf("UpdateRequestScript: no request '%s' found in collection '%s'", requestName, coll.Name)
	}
	req.Script = data.ScriptFromString(strings.Join(newScript, "\n"))
	return coll.UpsertRequest(req).Flush()
}

func SetCurrentEnv(newEnv string) error {
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		return err
	}
	conf.CurrentEnv = newEnv
	return conf.Flush()
}
