package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const defaultPerms = 0644

var CurrentVersion = NewSemVer(0, 1, 0)

type Squmpfile struct {
	Path        string    `json:"-"`
	Version     SemVer    `json:"version"`
	Title       string    `json:"title"`
	Requests    []Request `json:"requests"`
	Environment EnvMap    `json:"environment"`
}

type Request struct {
	Title  string `json:"title"`
	Script Script `json:"script"`
}

type Script []string

func (s Script) String() string {
	return strings.Join(s, "\n")
}

func ScriptFromString(s string) Script {
	return strings.Split(s, "\n")
}

func NewRequest(title string) *Request {
	return &Request{
		Title:  title,
		Script: Script{"print('hello world!')"},
	}
}

type EnvMap map[string]EnvMapValue

type EnvMapValue map[string]string

func (e EnvMap) PrintInfo() {
	fmt.Println("Environment:")
	if len(e) == 0 {
		fmt.Println("  <none>")
		return
	}
	for env, vars := range e {
		fmt.Printf("  %s\n", env)
		for k, v := range vars {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}
}

type SemVer struct {
	Major uint `json:"major"`
	Minor uint `json:"minor"`
	Patch uint `json:"patch"`
}

func NewSemVer(major, minor, patch uint) SemVer {
	return SemVer{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

func (s SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}

func (s SemVer) GreaterThan(other SemVer) bool {
	return s.Major > other.Major || s.Minor > other.Minor || s.Patch > other.Patch
}

func DefaultSqumpFile() Squmpfile {
	return Squmpfile{
		Path:     "Squmpfile.json",
		Version:  CurrentVersion,
		Title:    "My_New_Squmpfile",
		Requests: []Request{*NewRequest("NewReq")},
		Environment: EnvMap{
			"staging": {
				"hello": "world",
			},
		},
	}
}

type ErrNotFound struct {
	MissingItem string
	Location    string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("no %s found at: %s", e.MissingItem, e.Location)
}
func (e ErrNotFound) Is(target error) bool {
	return reflect.TypeOf(target) == reflect.TypeOf(ErrNotFound{})
}

func (s *Squmpfile) Flush() error {
	err := s.validate()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, b, defaultPerms)
}

func ReadSqumpfile(path string) (*Squmpfile, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound{
			MissingItem: "Squmpfile",
			Location:    path,
		}
	} else if err != nil {
		return nil, err
	}

	var s Squmpfile
	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}
	s.Path = path

	return &s, nil
}

func WriteDefaultSqumpfile() error {
	sf := DefaultSqumpFile()
	return sf.Flush()
}

func (s *Squmpfile) ExecuteRequest(
	conf *Config,
	reqName string,
	loopCheck LoopChecker,
	overrides EnvMapValue,
) (*State, error) {
	req, ok := s.GetRequest(reqName)
	if !ok {
		return nil, ErrNotFound{
			MissingItem: "request",
			Location:    reqName,
		}
	}

	return ExecuteRequest(
		conf,
		Identifier{
			Path:      s.Path,
			Squmpfile: s.Title,
			Request:   reqName,
		},
		req.Script.String(),
		s.Environment,
		overrides,
		loopCheck,
	)
}

func (s *Squmpfile) PrepareScript(conf *Config, reqName string, overrides EnvMapValue) (string, map[string]string, error) {
	req, ok := s.GetRequest(reqName)
	if !ok {
		return "", nil, ErrNotFound{
			MissingItem: "request",
			Location:    reqName,
		}
	}

	return PrepareScript(
		conf,
		Identifier{
			Path:      s.Path,
			Squmpfile: s.Title,
			Request:   reqName,
		},
		req.Script.String(),
		s.Environment,
		overrides,
	)
}

func (s *Squmpfile) GetRequest(req string) (*Request, bool) {
	for _, r := range s.Requests {
		if r.Title == req {
			return &r, true
		}
	}
	return nil, false
}

func (s *Squmpfile) RemoveRequest(title string) error {
	for i, r := range s.Requests {
		if r.Title == title {
			s.Requests = append(s.Requests[:i], s.Requests[i+1:]...)
			return s.Flush()
		}
	}
	return fmt.Errorf("no request titled '%s' found in squmpfile '%s'", title, s.Title)
}

func (s *Squmpfile) UpsertRequest(req *Request) *Squmpfile {
	found := false
	for i, r := range s.Requests {
		if r.Title == req.Title {
			s.Requests[i] = *req
			found = true
		}
	}
	if !found {
		s.Requests = append(s.Requests, *req)
	}
	return s
}

func (s *Squmpfile) EditRequest(reqName string) error {
	path := s.Path
	cb := func(b []byte) error {
		sq, err := ReadSqumpfile(path)
		if err != nil {
			return err
		}
		req, ok := sq.GetRequest(reqName)
		if !ok {
			return ErrNotFound{
				MissingItem: reqName,
				Location:    sq.Title,
			}
		}
		req.Script = ScriptFromString(strings.TrimSpace(string(b)))
		return sq.UpsertRequest(req).Flush()
	}

	req, ok := s.GetRequest(reqName)
	if !ok {
		return ErrNotFound{
			MissingItem: reqName,
			Location:    s.Title,
		}
	}
	b, err := EditBuffer([]byte(req.Script.String()), fmt.Sprintf("%s-%s-*.lua", s.Title, reqName), cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (s *Squmpfile) EditEnv() error {
	path := s.Path
	cb := func(b []byte) error {
		sq, err := ReadSqumpfile(path)
		if err != nil {
			return err
		}
		var e EnvMap
		err = json.Unmarshal(b, &e)
		if err != nil {
			return err
		}

		sq.Environment = e
		return sq.Flush()
	}

	envBytes, err := json.MarshalIndent(s.Environment, "", "  ")
	if err != nil {
		return err
	}

	basename := filepath.Base(s.Title)
	b, err := EditBuffer(envBytes, fmt.Sprintf("%s-config-*.json", basename), cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (s *Squmpfile) EditTitle() error {
	path := s.Path
	cb := func(b []byte) error {
		sq, err := ReadSqumpfile(path)
		if err != nil {
			return err
		}
		sq.Title = strings.TrimSpace(string(b))
		return sq.Flush()
	}

	basename := filepath.Base(s.Title)
	b, err := EditBuffer([]byte(s.Title), fmt.Sprintf("%s-title-*.json", basename), cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (s *Squmpfile) validate() error {
	if strings.Contains(s.Title, ".") {
		return fmt.Errorf("Illegal character '.' detected in squmpfile title: '%s'", s.Title)
	}
	for _, req := range s.Requests {
		if strings.Contains(req.Title, ".") {
			return fmt.Errorf("Illegal character '.' detected in request title: '%s'", req.Title)
		}
	}
	return nil
}

func (s *Squmpfile) SetEnvVar(env, key, val string) {
	if s.Environment == nil {
		s.Environment = make(EnvMap)
	}
	if s.Environment[env] == nil {
		s.Environment[env] = make(map[string]string)
	}
	s.Environment[env][key] = val
}

func (s *Squmpfile) PrintInfo() {
	strOrNone := func(s string) string {
		if s == "" {
			return "<none>"
		}
		return s
	}

	fmt.Println("Title:", strOrNone(s.Title))
	fmt.Println("Version:", strOrNone(s.Version.String()))
	fmt.Println("Requests:")
	for _, req := range s.Requests {
		fmt.Printf("  %s\n", strOrNone(req.Title))
	}
	s.Environment.PrintInfo()
}
