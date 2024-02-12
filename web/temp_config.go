package web

import (
	"net/http"
	"sync"

	"github.com/EvWilson/sqump/data"
)

type TempConfig struct {
	data.EnvMap
	sync.RWMutex
}

var tempConfig TempConfig

func saveTempConfig(req *http.Request) error {
	envMap, err := configMap(req)
	if err != nil {
		return err
	}
	tempConfig.Lock()
	defer tempConfig.Unlock()
	tempConfig.EnvMap = envMap
	return nil
}

func getTempConfig() data.EnvMap {
	tempConfig.RLock()
	defer tempConfig.RUnlock()
	return tempConfig.EnvMap.DeepCopy()
}
