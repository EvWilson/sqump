package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func EditBuffer(data []byte, tmpFilepattern string) ([]byte, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, errors.New("no value found in shell environment for EDITOR")
	}

	f, err := os.CreateTemp("", tmpFilepattern)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Printf("error closing file '%s': %v\n", file.Name(), err)
			return
		}
	}(f)
	defer func(filename string) {
		err = os.Remove(filename)
		if err != nil {
			fmt.Printf("error removing tmpfile '%s': %v\n", filename, err)
			return
		}
	}(f.Name())

	target := len(data)
	var current int64 = 0
	for {
		n, err := f.WriteAt(data, current)
		if err != nil {
			return nil, err
		}
		current += int64(n)
		if n == target {
			break
		}
	}

	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return b, nil
}
