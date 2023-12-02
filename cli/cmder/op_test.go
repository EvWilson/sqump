package cmder

import (
	"reflect"
	"strings"
	"testing"
)

func TestMappingExtract(t *testing.T) {
	strToArr := func(str string) []string {
		return strings.Split(str, " ")
	}

	t.Run("Basic", func(t *testing.T) {
		testStr := "progname arg1 -e test=one -e other=two arg2"
		testArr := strToArr(testStr)
		testArr, m, err := ExtractOverrideMappings(testArr, "-e")
		assert(t, err == nil, err)
		expected := map[string]string{
			"test":  "one",
			"other": "two",
		}
		assert(t, reflect.DeepEqual(m, expected), "map comparison", m, expected, testArr)
		assert(t, reflect.DeepEqual(testArr, strToArr("progname arg1 arg2")), "str comparison", testArr)
	})

	t.Run("End position flag", func(t *testing.T) {
		testStr := "progname arg1 -e test=one -e other=two arg2 -e"
		testArr := strings.Split(testStr, " ")
		testArr, m, err := ExtractOverrideMappings(testArr, "-e")
		assert(t, err == nil, err)
		expected := map[string]string{
			"test":  "one",
			"other": "two",
		}
		assert(t, reflect.DeepEqual(m, expected), "map comparison", m, expected, testArr)
		assert(t, reflect.DeepEqual(testArr, strToArr("progname arg1 arg2")), "str comparison", testArr)
	})
}

func assert(t *testing.T, value bool, args ...any) {
	if !value {
		t.Fatal(args...)
	}
}
