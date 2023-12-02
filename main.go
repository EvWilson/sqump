package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EvWilson/sqump/cli"
	"github.com/EvWilson/sqump/core"
)

func main() {
	// Get config
	_, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if errors.Is(err, core.ErrNotFound{}) {
		err = offerDefaultConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	root := cli.BuildRoot()
	err = root.Handle(os.Args[1:])
	if err != nil && err.Error() != "abort" {
		root.PrintUsage()
		fmt.Println("error while handling:", err)
		os.Exit(1)
	}
}

func offerDefaultConfig() error {
	fmt.Println("No config file detected. Would you like to create a new one? [Y/n]")
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	if s.Err() != nil {
		return s.Err()
	}
	if strings.ToLower(s.Text()) != "y" && s.Text() != "" {
		fmt.Println("Understood, have a nice day")
		os.Exit(0)
	}
	conf, err := core.CreateNewConfigFileAt(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	err = conf.Flush()
	if err != nil {
		return err
	}
	return nil
}
