package stores

import (
	"net/http"
	"sync"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/web/util"
	"github.com/gofrs/uuid/v5"
)

type TempConfigService interface {
	SaveTempConfig(req *http.Request) error
	GetTempEnv(req *http.Request) (data.EnvMap, error)
	GetTempEnvValue(req *http.Request) (data.EnvMapValue, error)
}

func NewTempConfigService(ces CurrentEnvService) TempConfigService {
	return &tempConfig{
		ces:     ces,
		envMap:  make(map[uuid.UUID]data.EnvMap),
		RWMutex: sync.RWMutex{},
	}
}

type tempConfig struct {
	ces    CurrentEnvService
	envMap map[uuid.UUID]data.EnvMap
	sync.RWMutex
}

func (t *tempConfig) SaveTempConfig(req *http.Request) error {
	uid, err := util.GetID(req)
	if err != nil {
		return err
	}
	envMap, err := util.ConfigMap(req)
	if err != nil {
		return err
	}
	t.Lock()
	defer t.Unlock()
	t.envMap[uid] = envMap
	return nil
}

func (t *tempConfig) GetTempEnv(req *http.Request) (data.EnvMap, error) {
	uid, err := util.GetID(req)
	if err != nil {
		return nil, err
	}
	t.RLock()
	defer t.RUnlock()
	if val, ok := t.envMap[uid]; ok {
		return val, nil
	} else {
		return make(data.EnvMap), nil
	}
}

func (t *tempConfig) GetTempEnvValue(req *http.Request) (data.EnvMapValue, error) {
	currentEnv, err := t.ces.GetCurrentEnv(req)
	if err != nil {
		return nil, err
	}
	envMap, err := t.GetTempEnv(req)
	if err != nil {
		return nil, err
	}
	if val, ok := envMap[currentEnv]; ok {
		return val, nil
	} else {
		return nil, nil
	}
}
