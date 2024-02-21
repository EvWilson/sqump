package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/EvWilson/sqump/data"
)

func ConfigMap(req *http.Request) (data.EnvMap, error) {
	configData, ok := req.Form["config"]
	if !ok {
		return nil, errors.New("core config form does not contain field 'config'")
	}
	configString := strings.Join(configData, "\n")
	var envMap data.EnvMap
	err := json.Unmarshal([]byte(configString), &envMap)
	if err != nil {
		return nil, err
	}
	return envMap, nil
}
