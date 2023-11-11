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
	envMap, ok := s.Environment[conf.CurrentEnv]
	if !ok {
		return "", fmt.Errorf("no environment found for name: %s", conf.CurrentEnv)
	}

	tmpl, err := template.New(s.Title).Parse(script)
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, envMap)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
