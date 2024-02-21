package stores

import (
	"net/http"
	"sync"

	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/web/util"
	"github.com/gofrs/uuid/v5"
)

type CurrentEnvService interface {
	GetCurrentEnv(r *http.Request) (string, error)
	SetCurrentEnv(r *http.Request, env string) error
}

func NewCurrentEnvService(isReadonlyMode bool) CurrentEnvService {
	return &currentEnvStore{
		isReadonlyMode: isReadonlyMode,
		envStore:       make(map[uuid.UUID]string),
		RWMutex:        sync.RWMutex{},
	}

}

type currentEnvStore struct {
	isReadonlyMode bool
	envStore       map[uuid.UUID]string
	sync.RWMutex
}

func (s *currentEnvStore) GetCurrentEnv(r *http.Request) (string, error) {
	uid, err := util.GetID(r)
	if err != nil {
		return "", err
	}
	if s.isReadonlyMode {
		s.RLock()
		defer s.RUnlock()
		if val, ok := s.envStore[uid]; ok {
			return val, nil
		} else {
			env, err := handlers.GetCurrentEnv()
			if err != nil {
				return "", err
			}
			return env, nil
		}
	} else {
		env, err := handlers.GetCurrentEnv()
		if err != nil {
			return "", err
		} else {
			return env, nil
		}
	}
}

func (s *currentEnvStore) SetCurrentEnv(r *http.Request, env string) error {
	uid, err := util.GetID(r)
	if err != nil {
		return err
	}
	if s.isReadonlyMode {
		s.Lock()
		defer s.Unlock()
		s.envStore[uid] = env
		return nil
	} else {
		return handlers.SetCurrentEnv(env)
	}
}
