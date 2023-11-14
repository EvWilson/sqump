package core

import (
	"bytes"
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
	ident Identifier,
	script string,
	env EnvMap,
	loopCheck LoopChecker,
) (*State, error) {
	mergedEnv, err := getMergedEnv(env)
	if err != nil {
		return nil, err
	}

	script, err = replaceEnvTemplates(ident.String(), script, mergedEnv)
	if err != nil {
		return nil, err
	}

	state := CreateState(ident, mergedEnv, loopCheck)
	defer state.Close()

	if err := state.DoString(script); err != nil {
		panic(err)
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

func getMergedEnv(squmpEnv EnvMap) (map[string]string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	squmpfileEnv, ok := squmpEnv[conf.CurrentEnv]
	if !ok {
		return nil, fmt.Errorf("no matching environment found in squmpfile for name: %s", conf.CurrentEnv)
	}
	configEnv, ok := conf.Environment[conf.CurrentEnv]
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
