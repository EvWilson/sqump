package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/EvWilson/sqump/test/example"
)

func main() {
	port := "5310"
	mux := example.MakeMux()
	fmt.Println("starting server at", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		fmt.Println("error while serving:", err)
		os.Exit(1)
	}
}
