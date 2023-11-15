package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/EvWilson/sqump/example"
)

func main() {
	port := "5309"
	mux := example.MakeMux()
	fmt.Println("starting server at", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		fmt.Println("error while serving:", err)
		os.Exit(1)
	}
}
