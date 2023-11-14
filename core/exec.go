package core

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"
)

type Identifier struct {
	Path      string
	Squmpfile string
	Request   string
}

func (i Identifier) String() string {
	return fmt.Sprintf("%s.%s.%s", i.Path, i.Squmpfile, i.Request)
}

func ExecuteRequest(
	conf *Config,
	ident Identifier,
	script string,
	env EnvMap,
	loopCheck LoopChecker,
) (*State, error) {
	mergedEnv, err := getMergedEnv(conf.CurrentEnv, env, conf.Environment)
	if err != nil {
		return nil, err
	}

	script, err = replaceEnvTemplates(ident.String(), script, mergedEnv)
	if err != nil {
		return nil, err
	}

	state := CreateState(conf, ident, mergedEnv, loopCheck)
	defer state.Close()

	err = state.DoString(script)
	if state.err != nil || err != nil {
		return nil, mergeErrors(state.err, err)
	}

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

func getMergedEnv(current string, squmpEnv, coreEnv EnvMap) (map[string]string, error) {
	squmpfileEnv, ok := squmpEnv[current]
	if !ok {
		return nil, fmt.Errorf("no matching environment found in squmpfile for name: %s", current)
	}
	configEnv, ok := coreEnv[current]
	if !ok {
		// Overrides in the core config are optional, so this is not a failure case
		configEnv = make(map[string]string)
	}
	return mergeMaps(squmpfileEnv, configEnv), nil
}

// mergeMaps will upsert later map entries into/over earlier map entries
func mergeMaps(m ...map[string]string) map[string]string {
	if len(m) == 0 {
		return make(map[string]string)
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
