package main

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/EvWilson/sqump/config"
	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/log"
	"github.com/EvWilson/sqump/web"
)

const port = "5309"

func main() {
	// Get config
	conf, err := core.ReadConfig()
	if errors.Is(err, core.ErrNotFound{}) {
		fmt.Println("No config file detected. Would you like to create a new one? [Y/n]")
		s := bufio.NewScanner(os.Stdin)
		s.Scan()
		if s.Err() != nil {
			die(s.Err())
		}
		if strings.ToLower(s.Text()) != "y" || s.Text() == "" {
			fmt.Println("Understood, have a nice day")
			os.Exit(0)
		}
		conf, err = core.CreateNewConfigFile()
		if err != nil {
			die(err)
		}
		err = conf.Flush()
		if err != nil {
			die(err)
		}
	} else if err != nil {
		die(err)
	}

	// Handle user command
	config.AssertMinArgLen(2, handlers.PrintUsage)
	cmd := os.Args[1]
	switch cmd {
	case "edit":
		config.AssertArgLen(4)
		handlers.HandleEdit(os.Args[2], os.Args[3])
	case "env":
		config.AssertMinArgLen(3, handlers.PrintUsage)
		err = handlers.HandleAllEnv(os.Args[2], os.Args)
		if err != nil {
			dieWithFunc(handlers.PrintUsage, err)
		}
	case "exec":
		config.AssertArgLen(4, handlers.PrintUsage)
		err := handlers.ExecuteRequest(os.Args[2], os.Args[3])
		if err != nil {
			die(err)
		}
	case "help":
		handlers.PrintUsage()
		return
	case "info":
		config.AssertMinArgLen(3, handlers.PrintUsage)
		handlers.HandleInfo(os.Args)
		return
	case "init":
		err := core.WriteDefaultSqumpfile()
		if err != nil {
			die(err)
		}
	case "register":
		err = conf.Register(cmd)
		if err != nil {
			die(err)
		}
	case "serve":
		config.AssertArgLen(2, handlers.PrintUsage)
		l := log.NewLogger(slog.LevelInfo)
		mux, err := web.NewRouter(l)
		if err != nil {
			die(err)
		}
		l.Info("starting server", "port", port)
		err = http.ListenAndServe(":"+port, mux)
		if err != nil {
			die(err)
		}
	default:
		dieWithFunc(handlers.PrintUsage, fmt.Errorf("handle error: unrecognized command: %s\n", cmd))
	}
}

func dieWithFunc(f func(), err error) {
	f()
	die(err)
}

func die(err error) {
	dieLeveled(err)
}

func dieLeveled(err error) {
	_, _, line, _ := runtime.Caller(2)
	fmt.Printf("error: line: %d: %v\n", line, err)
	os.Exit(1)
}
