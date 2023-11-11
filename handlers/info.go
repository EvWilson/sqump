package handlers

import (
	"fmt"

	"github.com/EvWilson/sqump/config"
	"github.com/EvWilson/sqump/core"
)

func HandleInfo(args []string) {
	switch args[2] {
	case "core":
		conf, err := core.ReadConfig()
		if err != nil {
			fmt.Println("error reading config:", err)
			return
		}
		conf.PrintInfo()
	case "file":
		config.AssertMinArgLen(4, PrintUsage)
		sq, err := core.ReadSqumpfile(args[3])
		if err != nil {
			fmt.Printf("error reading squmpfile at %s: %v\n", args[3], err)
			return
		}
		sq.PrintInfo()
	}
}
