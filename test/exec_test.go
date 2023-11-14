package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/example"
)

func TestBasicExample(t *testing.T) {
	configPath := "testdata/test_example_config.json"
	filePath := "testdata/test_example_squmpfile.json"

	mux := example.MakeMux()
	go func() {
		err := http.ListenAndServe(":5309", mux)
		if err != nil {
			fmt.Println("error from mux termination:", err)
		}
	}()

	conf, err := core.ReadConfigFrom(configPath)
	if err != nil {
		t.Fatal(err)
	}
	err = conf.Register(filePath)
	if err != nil {
		t.Fatal(err)
	}

	sq, err := core.ReadSqumpfile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sq.ExecuteRequest(conf, "A", make(core.LoopChecker))
	if err != nil {
		t.Fatal(err)
	}
}
