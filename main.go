package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/handlers/cmder"
)

func main() {
	// Get config
	_, err := core.ReadConfig()
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

	root := cmder.NewRoot("Welcome to sqump!", os.Stdout)
	root.Register(
		handlers.EditOperation(),
		handlers.ExecOperation(),
		handlers.InfoOperation(),
		cmder.NewOp(
			"init",
			"init",
			"Create a new default squmpfile in the current directory",
			func(args []string) error {
				return core.WriteDefaultSqumpfile()
			},
		),
		handlers.WebOperation(),
	)
	err = root.Handle(os.Args[1:])
	if err != nil {
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
	conf, err := core.CreateNewConfigFile()
	if err != nil {
		return err
	}
	err = conf.Flush()
	if err != nil {
		return err
	}
	return nil
}
