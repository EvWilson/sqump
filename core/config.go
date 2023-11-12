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
	Version     SemVer       `json:"version"`
	Files       []ConfigFile `json:"files"`
	CurrentEnv  string       `json:"current_env"`
	Environment EnvMap       `json:"environment"`
}

type ConfigFile struct {
	Path string `json:"path"`
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
	c.Files = append(c.Files, ConfigFile{
		Path: fullpath,
	})
	err = c.Flush()
	return err
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
		Files:       []ConfigFile{},
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
	for _, file := range c.Files {
		fmt.Printf("  %s\n", file.Path)
	}
	c.Environment.PrintInfo()
}

func AssertArgLen(expectedLen int, errFuncs ...func()) {
	_, file, line, _ := runtime.Caller(1)
	if len(os.Args) != expectedLen {
		fmt.Printf("error: %s:%d: expected %d arguments, received %d\n", file, line, expectedLen, len(os.Args))
		for _, f := range errFuncs {
			f()
		}
		os.Exit(1)
	}
}

func AssertMinArgLen(minLen int, errFuncs ...func()) {
	_, file, line, _ := runtime.Caller(1)
	if len(os.Args) < minLen {
		fmt.Printf("error: %s:%d: expected at least %d arguments, received %d\n", file, line, minLen, len(os.Args))
		for _, f := range errFuncs {
			f()
		}
		os.Exit(1)
	}
}
