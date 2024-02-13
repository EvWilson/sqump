package main

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/EvWilson/sqump/cli"
	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/prnt"
)

func main() {
	prnt.SetPrinter(&prnt.StandardPrinter{})
	// Get config
	_, err := handlers.GetConfig()
	if errors.Is(err, data.ErrNotFound{}) {
		err = offerDefaultConfig()
		if err != nil {
			prnt.Println(err)
			os.Exit(1)
		}
	} else if err != nil {
		prnt.Println(err)
		os.Exit(1)
	}

	root := cli.BuildRoot()
	err = root.Handle(os.Args[1:])
	if err != nil && err.Error() != "abort" {
		root.PrintUsage()
		prnt.Println("error while handling:", err)
		os.Exit(1)
	}
}

func offerDefaultConfig() error {
	prnt.Println("No config file detected. Would you like to create a new one? [Y/n]")
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	if s.Err() != nil {
		return s.Err()
	}
	if strings.ToLower(s.Text()) != "y" && s.Text() != "" {
		prnt.Println("Understood, have a nice day")
		os.Exit(0)
	}
	conf, err := data.CreateNewConfigFileAt(data.DefaultConfigLocation())
	if err != nil {
		return err
	}
	err = conf.Flush()
	if err != nil {
		return err
	}
	return nil
}
