package core

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func (s *Squmpfile) EditRequest(squmpFilePath, reqName string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("no value found in shell environment for EDITOR")
	}

	req, ok := s.GetRequest(reqName)
	if !ok {
		return ErrNotFound{
			MissingItem: reqName,
			Location:    s.Title,
		}
	}

	f, err := os.CreateTemp("", fmt.Sprintf("%s-%s-*.lua", s.Title, reqName))
	if err != nil {
		return err
	}
	defer func(filename string) {
		os.Remove(filename)
	}(f.Name())
	target, data := len(req.Script), []byte(req.Script)
	var current int64 = 0
	for {
		n, err := f.WriteAt(data, current)
		if err != nil {
			return err
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
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	req.Script = string(b)
	err = s.UpsertRequest(req).Flush(squmpFilePath)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return nil
	}

	return nil
}
