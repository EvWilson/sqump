package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/EvWilson/sqump/prnt"
)

const defaultPerms = 0644

var (
	CurrentVersion = NewSemVer(0, 1, 0)
	collLock       = make(map[string]*sync.RWMutex, 0)
)

func collRLock(path string) func() {
	lock, ok := collLock[path]
	if !ok {
		lock = &sync.RWMutex{}
		collLock[path] = lock
	}
	lock.RLock()
	return lock.RUnlock
}

func collWLock(path string) func() {
	lock, ok := collLock[path]
	if !ok {
		lock = &sync.RWMutex{}
		collLock[path] = lock
	}
	lock.Lock()
	return lock.Unlock
}

type Collection struct {
	Path        string    `json:"-"`
	Version     SemVer    `json:"version"`
	Name        string    `json:"name"`
	Requests    []Request `json:"requests"`
	Environment EnvMap    `json:"environment"`
}

type Request struct {
	Name   string `json:"name"`
	Script Script `json:"script"`
}

type Script []string

func (s Script) String() string {
	return strings.Join(s, "\n")
}

func ScriptFromString(s string) Script {
	split := strings.Split(s, "\n")
	// Try to defend against \r\n
	for i, sp := range split {
		split[i] = strings.TrimSuffix(sp, "\r")
	}
	return split
}

func NewRequest(name string) *Request {
	return &Request{
		Name: name,
		Script: Script{
			"local s = require('sqump')",
			"",
			"local resp = s.fetch('http://localhost:8000')",
			"",
			"s.print_response(resp)",
		},
	}
}

type EnvMap map[string]EnvMapValue

func (em EnvMap) validate() error {
	for k, v := range em {
		if strings.Contains(k, "-") {
			return fmt.Errorf("cannot accept map key '%s' with '-' character", k)
		}
		if err := v.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (em EnvMap) DeepCopy() EnvMap {
	ret := make(EnvMap, len(em))
	for k, v := range em {
		val := make(EnvMapValue, len(v))
		for valK, valV := range v {
			val[valK] = valV
		}
		ret[k] = val
	}
	return ret
}

type EnvMapValue map[string]string

func (emv EnvMapValue) validate() error {
	for k := range emv {
		if strings.Contains(k, "-") {
			return fmt.Errorf("cannot accept submap key '%s' with '-' character", k)
		}
	}
	return nil
}

func (e EnvMap) PrintInfo() {
	prnt.Println("Environment:")
	if len(e) == 0 {
		prnt.Println("  <none>")
		return
	}
	for env, vars := range e {
		prnt.Printf("  %s\n", env)
		for k, v := range vars {
			prnt.Printf("    %s: %s\n", k, v)
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

func DefaultCollection() Collection {
	return Collection{
		Path:     "Squmpfile.json",
		Version:  CurrentVersion,
		Name:     "My_New_Collection",
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

func (c *Collection) Flush() error {
	unlock := collWLock(c.Path)
	defer unlock()
	err := c.validate()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.Path, b, defaultPerms)
}

func ReadCollection(path string) (*Collection, error) {
	unlock := collRLock(path)
	defer unlock()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound{
			MissingItem: "collection",
			Location:    path,
		}
	} else if err != nil {
		return nil, err
	}

	var s Collection
	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}
	s.Path = path

	return &s, nil
}

func WriteDefaultCollection() error {
	sf := DefaultCollection()
	return sf.Flush()
}

func (c *Collection) GetRequest(req string) (*Request, bool) {
	for _, r := range c.Requests {
		if r.Name == req {
			return &r, true
		}
	}
	return nil, false
}

func (c *Collection) RemoveRequest(name string) error {
	for i, r := range c.Requests {
		if r.Name == name {
			c.Requests = append(c.Requests[:i], c.Requests[i+1:]...)
			return c.Flush()
		}
	}
	return fmt.Errorf("no request named '%s' found in collection '%s'", name, c.Name)
}

func (c *Collection) UpsertRequest(req *Request) *Collection {
	found := false
	for i, r := range c.Requests {
		if r.Name == req.Name {
			c.Requests[i] = *req
			found = true
		}
	}
	if !found {
		c.Requests = append(c.Requests, *req)
	}
	return c
}

func (c *Collection) EditRequest(reqName string) error {
	path := c.Path
	cb := func(b []byte) error {
		coll, err := ReadCollection(path)
		if err != nil {
			return err
		}
		req, ok := coll.GetRequest(reqName)
		if !ok {
			return ErrNotFound{
				MissingItem: reqName,
				Location:    coll.Name,
			}
		}
		req.Script = ScriptFromString(strings.TrimSpace(string(b)))
		return coll.UpsertRequest(req).Flush()
	}

	req, ok := c.GetRequest(reqName)
	if !ok {
		return ErrNotFound{
			MissingItem: reqName,
			Location:    c.Name,
		}
	}
	b, err := EditBuffer([]byte(req.Script.String()), fmt.Sprintf("%s-%s-*.lua", c.Name, reqName), cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (c *Collection) EditEnv() error {
	path := c.Path
	cb := func(b []byte) error {
		coll, err := ReadCollection(path)
		if err != nil {
			return err
		}
		var e EnvMap
		err = json.Unmarshal(b, &e)
		if err != nil {
			return err
		}

		coll.Environment = e
		return coll.Flush()
	}

	envBytes, err := json.MarshalIndent(c.Environment, "", "  ")
	if err != nil {
		return err
	}

	basename := filepath.Base(c.Name)
	b, err := EditBuffer(envBytes, fmt.Sprintf("%s-config-*.json", basename), cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (c *Collection) EditName() error {
	path := c.Path
	cb := func(b []byte) error {
		coll, err := ReadCollection(path)
		if err != nil {
			return err
		}
		coll.Name = strings.TrimSpace(string(b))
		return coll.Flush()
	}

	basename := filepath.Base(c.Name)
	b, err := EditBuffer([]byte(c.Name), fmt.Sprintf("%s-name-*.json", basename), cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (c *Collection) validate() error {
	if err := c.Environment.validate(); err != nil {
		return err
	}
	if strings.Contains(c.Name, ".") {
		return fmt.Errorf("Illegal character '.' detected in collection name '%s'", c.Name)
	}
	reqNames := make(map[string]bool, len(c.Requests))
	for _, req := range c.Requests {
		if strings.Contains(req.Name, ".") {
			return fmt.Errorf("Illegal character '.' detected in request name '%s'", req.Name)
		}
		if _, ok := reqNames[req.Name]; ok {
			return fmt.Errorf("duplicate request name '%s'", req.Name)
		} else {
			reqNames[req.Name] = true
		}
	}
	return nil
}

func (c *Collection) SetEnvVar(env, key, val string) {
	if c.Environment == nil {
		c.Environment = make(EnvMap)
	}
	if c.Environment[env] == nil {
		c.Environment[env] = make(map[string]string)
	}
	c.Environment[env][key] = val
}

func (c *Collection) PrintInfo() {
	strOrNone := func(s string) string {
		if s == "" {
			return "<none>"
		}
		return s
	}

	prnt.Println("Name:", strOrNone(c.Name))
	prnt.Println("Version:", strOrNone(c.Version.String()))
	prnt.Println("Requests:")
	for _, req := range c.Requests {
		prnt.Printf("  %s\n", strOrNone(req.Name))
	}
	c.Environment.PrintInfo()
}
