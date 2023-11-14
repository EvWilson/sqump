package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/EvWilson/sqump/example"
)

func main() {
	mux := example.MakeMux()
	fmt.Println("starting server at 5000")
	err := http.ListenAndServe(":5000", mux)
	if err != nil {
		fmt.Println("error while serving:", err)
		os.Exit(1)
	}
}
