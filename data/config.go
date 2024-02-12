package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/EvWilson/sqump/prnt"
)

// DefaultConfigLocation returns the location of the sqump config file
func DefaultConfigLocation() string {
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "sqump", "config.json")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), ".config", "sqump", "config.json")
	case "windows":
		// todo: this needs testing
		return filepath.Join(os.Getenv("{FOLDERID_LocalAppData}"), "sqump", "config")
	default:
		panic("encountered unsupported GOOS platform:" + runtime.GOOS)
	}
}

type Config struct {
	Path        string   `json:"-"`
	Version     SemVer   `json:"version"`
	Files       []string `json:"files"`
	CurrentEnv  string   `json:"current_env"`
	Environment EnvMap   `json:"environment"`
}

func ReadConfigFrom(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound{
			MissingItem: "config file",
			Location:    path,
		}
	} else if err != nil {
		return nil, err
	}

	var c Config
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	c.Path = path

	return &c, nil
}

func (c *Config) Flush() error {
	if err := c.validate(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(c.Path, b, defaultPerms)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) validate() error {
	if err := c.Environment.validate(); err != nil {
		return err
	}
	return nil
}

func (c *Config) Register(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, it must be a file", path)
	}
	fullpath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	c.Files = append(c.Files, fullpath)
	err = c.Flush()
	return err
}

func (c *Config) Unregister(path string) error {
	for i, registered := range c.Files {
		if registered == path {
			c.Files = append(c.Files[:i], c.Files[i+1:]...)
			return c.Flush()
		}
	}

	return fmt.Errorf("no registered collection found for path '%s'", path)
}

func (c *Config) CheckForRegisteredFile(path string) (bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(abs)
	if err != nil {
		return false, err
	}
	for _, fpath := range c.Files {
		if abs == fpath {
			return true, nil
		}
	}
	return false, nil
}

func CreateNewConfigFileAt(path string) (*Config, error) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}
	_, err = os.Create(path)
	if err != nil {
		return nil, err
	}

	c := DefaultConfig(path)
	err = c.Flush()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func DefaultConfig(path string) *Config {
	return &Config{
		Path:        path,
		CurrentEnv:  "staging",
		Version:     CurrentVersion,
		Files:       []string{},
		Environment: EnvMap{},
	}
}

func (c *Config) EditEnv() error {
	path := c.Path
	cb := func(b []byte) error {
		conf, err := ReadConfigFrom(path)
		if err != nil {
			return err
		}

		var e EnvMap
		err = json.Unmarshal(b, &e)
		if err != nil {
			return err
		}

		conf.Environment = e
		return conf.Flush()
	}

	envBytes, err := json.MarshalIndent(c.Environment, "", "  ")
	if err != nil {
		return err
	}
	b, err := EditBuffer(envBytes, "core-config-*.json", cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (c *Config) EditCurrentEnv() error {
	path := c.Path
	cb := func(b []byte) error {
		conf, err := ReadConfigFrom(path)
		if err != nil {
			return err
		}
		conf.CurrentEnv = strings.TrimSpace(string(b))
		return conf.Flush()
	}

	b, err := EditBuffer([]byte(c.CurrentEnv), "core-config-current-env-*.json", cb)
	if err != nil {
		return err
	}

	return cb(b)
}

func (c *Config) CollectionByName(name string) (*Collection, error) {
	for _, fpath := range c.Files {
		sq, err := ReadCollection(fpath)
		if err != nil {
			return nil, err
		}
		if sq.Name == name {
			return sq, nil
		}
	}
	return nil, fmt.Errorf("no collection found for name '%s'", name)
}

func (c *Config) PrintInfo() {
	strOrNone := func(s string) string {
		if s == "" {
			return "<none>"
		}
		return s
	}

	prnt.Println("Current Env:", strOrNone(c.CurrentEnv))
	prnt.Println("Version:", strOrNone(c.Version.String()))
	prnt.Println("Files:")
	if len(c.Files) == 0 {
		prnt.Println("  <none>")
		return
	}
	for _, fpath := range c.Files {
		prnt.Printf("  %s\n", fpath)
	}
	c.Environment.PrintInfo()
}
