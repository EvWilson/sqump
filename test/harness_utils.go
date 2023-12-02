package test

import (
	"os"
	"testing"
)

type Tmpfile struct {
	F *os.File
}

func CreateTmpfile(path string) (*Tmpfile, error) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(f.Name(), b, info.Mode())
	if err != nil {
		return nil, err
	}
	return &Tmpfile{
		F: f,
	}, nil
}

func (t *Tmpfile) Cleanup() error {
	return os.Remove(t.F.Name())
}

func assert(t *testing.T, value bool, args ...any) {
	if !value {
		t.Fatal(args...)
	}
}
