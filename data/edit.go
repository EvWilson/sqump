package data

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/EvWilson/sqump/prnt"
	"github.com/fsnotify/fsnotify"
)

func EditBuffer(
	data []byte,
	tmpFilepattern string,
	saveCallback func(b []byte) error,
) ([]byte, error) {
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
			prnt.Printf("error closing file '%s': %v\n", file.Name(), err)
			return
		}
	}(f)
	defer func(filename string) {
		err = os.Remove(filename)
		if err != nil {
			prnt.Printf("error removing tmpfile '%s': %v\n", filename, err)
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

	cmdCtx, cmdCancel := context.WithCancel(context.Background())
	defer cmdCancel()
	cmd := exec.CommandContext(cmdCtx, editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	w, err := watcher(f.Name(), done, saveCallback)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func watcher(
	path string,
	waitChan chan error,
	saveCallback func(b []byte) error,
) (*fsnotify.Watcher, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err

	}

	// watching directory because: https://github.com/fsnotify/fsnotify#watching-a-file-doesnt-work-well
	err = watcher.Add(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil, errors.New("bailed on watcher event")
			}
			if event.Has(fsnotify.Write) {
				if event.Name == abspath {
					b, err := os.ReadFile(abspath)
					if err != nil {
						return nil, err
					}
					err = saveCallback(b)
					if err != nil {
						return nil, err
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil, fmt.Errorf("bailed on watcher error: %v", err)
			}
			return nil, fmt.Errorf("watcher error: %v", err)
		case <-waitChan:
			return watcher, nil
		}
	}
}
