package exec

import (
	"strconv"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestStateHelpers(t *testing.T) {
	t.Run("Test sliceToLuaArray", func(t *testing.T) {
		goSlice := []string{"a", "b", "c", "d"}
		arr := sliceToLuaArray(goSlice)
		arr.ForEach(func(l1, l2 lua.LValue) {
			idx, err := strconv.Atoi(l1.String())
			if err != nil {
				t.Fatal(err)
			}
			val1 := goSlice[idx-1]
			if l2.String() != val1 {
				t.Fatalf("'%s' does not equal '%v'", l2.String(), val1)
			}
		})
	})
}
