package exec

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/prnt"
)

type Identifier struct {
	Path       string
	Collection string
	Request    string
}

func (i Identifier) String() string {
	return fmt.Sprintf("%s.%s.%s", i.Path, i.Collection, i.Request)
}

func PrepareScript(
	sq *data.Collection,
	requestName string,
	conf *data.Config,
	overrides data.EnvMapValue,
) (string, data.EnvMapValue, error) {
	req, ok := sq.GetRequest(requestName)
	if !ok {
		return "", nil, data.ErrNotFound{
			MissingItem: "request",
			Location:    requestName,
		}
	}
	ident := Identifier{
		Path:       sq.Path,
		Collection: sq.Name,
		Request:    requestName,
	}
	return prepScript(conf, ident, req.Script.String(), sq.Environment, overrides)
}

func prepScript(
	conf *data.Config,
	ident Identifier,
	script string,
	requestEnv data.EnvMap,
	overrides data.EnvMapValue,
) (string, data.EnvMapValue, error) {
	mergedEnv, err := getMergedEnv(conf.CurrentEnv, requestEnv, conf.Environment, overrides)
	if err != nil {
		return "", nil, err
	}

	script, err = replaceEnvTemplates(ident.String(), script, mergedEnv)
	if err != nil {
		return "", nil, err
	}
	return script, mergedEnv, nil
}

func ExecuteRequest(
	sq *data.Collection,
	requestName string,
	conf *data.Config,
	overrides data.EnvMapValue,
	loopCheck LoopChecker,
) (*State, error) {
	req, ok := sq.GetRequest(requestName)
	if !ok {
		return nil, data.ErrNotFound{
			MissingItem: "request",
			Location:    requestName,
		}
	}
	ident := Identifier{
		Path:       sq.Path,
		Collection: sq.Name,
		Request:    requestName,
	}

	script, mergedEnv, err := prepScript(conf, ident, req.Script.String(), sq.Environment, overrides)
	if err != nil {
		return nil, err
	}

	state := CreateState(conf, ident, mergedEnv, loopCheck)
	defer state.Close()

	err = state.DoString(script)
	if state.err != nil || err != nil {
		return nil, mergeErrors(state.err, err)
	}
	prnt.Printf("<script '%s: %s' complete>\n", sq.Name, requestName)

	return state, nil
}

// replaceEnvTemplates takes a script body and inserts environment
// data into template placeholders
func replaceEnvTemplates(ident, script string, env map[string]string) (string, error) {
	tmpl, err := template.New(ident).Parse(script)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, env)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getMergedEnv(
	current string,
	squmpEnv,
	coreEnv data.EnvMap,
	overrides data.EnvMapValue,
) (data.EnvMapValue, error) {
	collectionEnv, ok := squmpEnv[current]
	if !ok {
		return nil, fmt.Errorf("no matching environment found in collection for name: %s", current)
	}
	configEnv, ok := coreEnv[current]
	if !ok {
		// Overrides in the core config are optional, so this is not a failure case
		configEnv = make(map[string]string)
	}
	return mergeMaps(collectionEnv, configEnv, overrides), nil
}

// mergeMaps will upsert later map entries into/over earlier map entries
func mergeMaps(m ...data.EnvMapValue) data.EnvMapValue {
	if len(m) == 0 {
		return make(data.EnvMapValue)
	}

	res := m[0]
	for _, other := range m[1:] {
		for k, v := range other {
			res[k] = v
		}
	}

	return res
}

func mergeErrors(errs ...error) error {
	res := ""
	for _, e := range errs {
		if e == nil {
			continue
		}
		res += e.Error() + "\n"
	}
	return errors.New(res)
}
