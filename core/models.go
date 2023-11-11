package core

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

const defaultPerms = 0644

var CurrentVersion = NewSemVer(0, 1, 0)

type Squmpfile struct {
	Version     SemVer    `json:"version"`
	Title       string    `json:"title"`
	Requests    []Request `json:"requests"`
	Environment EnvMap    `json:"environment"`
}

type Request struct {
	Title  string `json:"title"`
	Script string `json:"script"`
}

type EnvMap map[string]map[string]string

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
		Version: CurrentVersion,
		Title:   "My New Squmpfile",
		Requests: []Request{
			{
				Title:  "NewReq",
				Script: "print('hello, world!')",
			},
		},
		Environment: EnvMap{},
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

func (s *Squmpfile) Flush(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, b, defaultPerms)
	if err != nil {
		return err
	}
	return nil
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

	return &s, nil
}

func WriteDefaultSqumpfile() error {
	sf := DefaultSqumpFile()
	err := sf.Flush("Squmpfile")
	if err != nil {
		return err
	}
	return nil
}

func (s *Squmpfile) GetRequest(req string) (*Request, bool) {
	for _, r := range s.Requests {
		if r.Title == req {
			return &r, true
		}
	}
	return nil, false
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
	fmt.Println("Environment:")
	if len(s.Environment) == 0 {
		fmt.Println("  <none>")
		return
	}
	for env, vars := range s.Environment {
		fmt.Printf("  %s\n", env)
		for k, v := range vars {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}
}
