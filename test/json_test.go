package test

import (
	"testing"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/exec"
)

func TestJSON(t *testing.T) {
	t.Run("to_json", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_json_squmpfile.json")
		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "to_json", conf, make(data.EnvMapValue), make(exec.LoopChecker))
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("from_json", func(t *testing.T) {
		tmpConf, tmpFile := setup(t, "testdata/test_example_config.json", "testdata/test_example_json_squmpfile.json")
		conf, err := data.ReadConfigFrom(tmpConf.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		coll, err := data.ReadCollection(tmpFile.F.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = exec.ExecuteRequest(coll, "from_json", conf, make(data.EnvMapValue), make(exec.LoopChecker))
		if err != nil {
			t.Fatal(err)
		}
	})
}
