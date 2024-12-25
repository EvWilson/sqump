package stores

import (
	"net/http"

	"github.com/EvWilson/sqump/handlers"
)

type ExecProxyService interface {
	ExecuteRequest(fpath, requestName string, r *http.Request) error
	GetPreparedScript(fpath, requestName string, r *http.Request) (string, error)
	CancelScripts()
}

func NewExecProxyService(ces CurrentEnvService, tcs TempConfigService) ExecProxyService {
	return &execProxyService{
		ces: ces,
		tcs: tcs,
	}
}

type execProxyService struct {
	ces CurrentEnvService
	tcs TempConfigService
}

func (e *execProxyService) ExecuteRequest(fpath, requestName string, r *http.Request) error {
	currentEnv, err := e.ces.GetCurrentEnv(r)
	if err != nil {
		return err
	}
	env, err := e.tcs.GetTempEnvValue(r)
	if err != nil {
		return err
	}
	return handlers.ExecuteRequest(fpath, requestName, currentEnv, env)
}

func (e *execProxyService) GetPreparedScript(fpath, requestName string, r *http.Request) (string, error) {
	currentEnv, err := e.ces.GetCurrentEnv(r)
	if err != nil {
		return "", err
	}
	env, err := e.tcs.GetTempEnvValue(r)
	if err != nil {
		return "", err
	}
	return handlers.GetPreparedScript(fpath, requestName, currentEnv, env)
}

func (e *execProxyService) CancelScripts() {
	handlers.CancelScripts()
}
