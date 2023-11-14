package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Location returns the location of the sqump config file
func Location() string {
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
	Version     SemVer   `json:"version"`
	Files       []string `json:"files"`
	CurrentEnv  string   `json:"current_env"`
	Environment EnvMap   `json:"environment"`
}

func ReadConfig() (*Config, error) {
	b, err := os.ReadFile(Location())
	if os.IsNotExist(err) {
		return nil, ErrNotFound{
			MissingItem: "config file",
			Location:    Location(),
		}
	} else if err != nil {
		return nil, err
	}

	var c Config
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) Flush() error {
	b, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(Location(), b, defaultPerms)
	if err != nil {
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

func (c *Config) CheckForRegisteredFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	_, err = os.Stat(abs)
	if err != nil {
		return err
	}
	for _, fpath := range c.Files {
		if abs == fpath {
			return nil
		}
	}
	return fmt.Errorf("error: filepath '%s' is not registered", abs)
}

func CreateNewConfigFile() (*Config, error) {
	err := os.MkdirAll(filepath.Dir(Location()), defaultPerms)
	if err != nil {
		return nil, err
	}
	_, err = os.Create(Location())
	if err != nil {
		return nil, err
	}

	c := DefaultConfig()
	err = c.Flush()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func DefaultConfig() *Config {
	return &Config{
		CurrentEnv:  "staging",
		Version:     CurrentVersion,
		Files:       []string{},
		Environment: EnvMap{},
	}
}

func (c *Config) EditEnv() error {
	envBytes, err := json.MarshalIndent(c.Environment, "", "  ")
	if err != nil {
		return err
	}

	b, err := EditBuffer(envBytes, "core-config-*.json")
	if err != nil {
		return err
	}

	var e EnvMap
	err = json.Unmarshal(b, &e)
	if err != nil {
		return err
	}

	c.Environment = e
	err = c.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) PrintInfo() {
	strOrNone := func(s string) string {
		if s == "" {
			return "<none>"
		}
		return s
	}

	fmt.Println("Current Env:", strOrNone(c.CurrentEnv))
	fmt.Println("Version:", strOrNone(c.Version.String()))
	fmt.Println("Files:")
	if len(c.Files) == 0 {
		fmt.Println("  <none>")
		return
	}
	for _, fpath := range c.Files {
		fmt.Printf("  %s\n", fpath)
	}
	c.Environment.PrintInfo()
}
