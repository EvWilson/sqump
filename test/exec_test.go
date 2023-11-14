package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/example"
)

func TestBasicExample(t *testing.T) {
	mux := example.MakeMux()
	go func() {
		err := http.ListenAndServe(":5309", mux)
		if err != nil {
			fmt.Println("error from mux termination:", err)
		}
	}()

	sq, err := core.ReadSqumpfile("testdata/test_example_squmpfile.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = sq.ExecuteRequest("A", make(core.LoopChecker))
	if err != nil {
		t.Fatal(err)
	}
}
