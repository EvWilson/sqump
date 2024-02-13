package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/exec"
	"github.com/EvWilson/sqump/prnt"
	"github.com/EvWilson/sqump/test/example"
)

func TestExample(t *testing.T) {
	prnt.SetPrinter(&prnt.StandardPrinter{})

	mux := example.MakeMux()
	go func() {
		err := http.ListenAndServe(":5310", mux)
		if err != nil {
			fmt.Println("error from mux termination:", err)
		}
	}()

	setup := func(t *testing.T, confPath, filePath string) (*Tmpfile, *Tmpfile) {
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
		_, err = exec.ExecuteRequest(coll, "A", conf, make(data.EnvMapValue), make(exec.LoopChecker))
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
		_, err = exec.ExecuteRequest(coll, "B", conf, make(data.EnvMapValue), make(exec.LoopChecker))
		if err != nil {
			t.Fatal(err)
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
		_, err = exec.ExecuteRequest(coll, "A", conf, data.EnvMapValue{
			"two": "2",
		}, make(exec.LoopChecker))
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Test JSON drilling", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_basic_squmpfile.json")
		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "C", conf, make(data.EnvMapValue), make(exec.LoopChecker))
		if err != nil {
			t.Fatal(err)
		}
	})
}
