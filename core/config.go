package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// configLocation returns the location of the sqump config file
func configLocation() string {
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
	return ReadConfigFrom(configLocation())
}

func ReadConfigFrom(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, ErrNotFound{
			MissingItem: "config file",
			Location:    configLocation(),
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
	return c.FlushTo(configLocation())
}

func (c *Config) FlushTo(path string) error {
	b, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, b, defaultPerms)
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
	err := os.MkdirAll(filepath.Dir(configLocation()), defaultPerms)
	if err != nil {
		return nil, err
	}
	_, err = os.Create(configLocation())
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

	cb := func(b []byte) error {
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

	b, err := EditBuffer(envBytes, "core-config-*.json", cb)
	if err != nil {
		return err
	}

	err = cb(b)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) SqumpfileByTitle(title string) (*Squmpfile, error) {
	for _, fpath := range c.Files {
		sq, err := ReadSqumpfile(fpath)
		if err != nil {
			return nil, err
		}
		if sq.Title == title {
			return sq, nil
		}
	}
	return nil, fmt.Errorf("no squmpfile found for title '%s'", title)
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
