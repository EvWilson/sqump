package test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/exec"
	"github.com/EvWilson/sqump/prnt"
	"github.com/EvWilson/sqump/test/example"
)

func setup(t *testing.T, confPath, filePath string) (*Tmpfile, *Tmpfile) {
	tmpConf, err := CreateTmpfile(confPath)
	assert(t, err == nil, "create conf")
	tmpFile, err := CreateTmpfile(filePath)
	assert(t, err == nil, "create file")
	t.Cleanup(func() {
		_ = tmpConf.Cleanup()
		assert(t, err == nil, "cleanup conf")
		_ = tmpFile.Cleanup()
		assert(t, err == nil, "cleanup file")
	})
	return tmpConf, tmpFile
}

func TestExample(t *testing.T) {
	prnt.SetPrinter(&prnt.StandardPrinter{})

	mux := example.MakeMux()
	go func() {
		err := http.ListenAndServe(":5310", mux)
		if err != nil {
			fmt.Println("error from mux termination:", err)
		}
	}()

	t.Run("Basic", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_basic_squmpfile.json")

		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		err = conf.Register(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}

		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "GetAuth", conf.CurrentEnv, make(data.EnvMapValue), exec.NewLoopChecker())
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Chained, gets result", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_basic_squmpfile.json")
		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "GetPayload", conf.CurrentEnv, make(data.EnvMapValue), exec.NewLoopChecker())
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Error on cyclical script execution", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_basic_squmpfile.json")
		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "Cycle1", conf.CurrentEnv, make(data.EnvMapValue), exec.NewLoopChecker())
		if err == nil || !strings.Contains(err.Error(), "cyclical") {
			t.Fatal("cyclical script execution did not produce proper error")
		}
	})

	t.Run("Multiple override sets and sources", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_multi_env_squmpfile.json")
		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "AssertFromEnvAndManualOverride", conf.CurrentEnv, data.EnvMapValue{
			"two": "2",
		}, exec.NewLoopChecker())
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Override in execute call", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_multi_env_squmpfile.json")
		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "AssertOnRequiredValue", conf.CurrentEnv, make(data.EnvMapValue), exec.NewLoopChecker())
		if err != nil {
			t.Fatal(err)
		}
	})
}
