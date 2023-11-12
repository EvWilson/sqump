package core

import (
	"bytes"
	"fmt"
	"text/template"
)

func (s *Squmpfile) ExecuteRequest(reqName string) error {
	req, ok := s.GetRequest(reqName)
	if !ok {
		return ErrNotFound{
			MissingItem: "request",
			Location:    reqName,
		}
	}

	script, err := s.ReplaceEnvTemplates(req.Script)
	if err != nil {
		return err
	}

	L := LoadState()
	defer L.Close()

	if err := L.DoString(script); err != nil {
		panic(err)
	}

	return nil
}

// ReplaceEnvTemplates takes a script body and inserts environment
// data into template placeholders
func (s *Squmpfile) ReplaceEnvTemplates(script string) (string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return "", err
	}
	squmpfileEnv, ok := s.Environment[conf.CurrentEnv]
	if !ok {
		return "", fmt.Errorf("no matching environment found in squmpfile '%s' for name: %s", s.Title, conf.CurrentEnv)
	}
	configEnv, ok := conf.Environment[conf.CurrentEnv]
	if !ok {
		// Overrides in the core config are optional, so this is not a failure case
		configEnv = make(map[string]string)
	}
	consolidated := mergeMaps(squmpfileEnv, configEnv)

	tmpl, err := template.New(s.Title).Parse(script)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, consolidated)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
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
