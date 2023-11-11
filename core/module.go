package core

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func LoadState() *lua.LState {
	L := lua.NewState(lua.Options{
		// SkipOpenLibs: true,
	})

	// for _, pair := range []struct {
	// 	n string
	// 	f lua.LGFunction
	// }{
	// 	{lua.LoadLibName, lua.OpenPackage}, // Must be first
	// 	{lua.BaseLibName, lua.OpenBase},
	// 	{lua.TabLibName, lua.OpenTable},
	// } {
	// 	if err := L.CallByParam(lua.P{
	// 		Fn:      L.NewFunction(pair.f),
	// 		NRet:    0,
	// 		Protect: true,
	// 	}, lua.LString(pair.n)); err != nil {
	// 		panic(err)
	// 	}
	// }

	L.PreloadModule("sqump", func(l *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"fetch": fetch,
		})
		L.Push(mod)
		return 1
	})

	return L
}

// https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API
func fetch(L *lua.LState) int {
	resource := L.ToString(1)
	options := L.ToTable(2)
	fmt.Println("resource:", resource, ", options:", options)

	method := stringOrDefault(options, "method", "GET")
	timeout := intOrDefault(options, "timeout", 10)

	req, err := http.NewRequest(method, resource, nil)
	if err != nil {
		fmt.Println("error creating request:", err)
		return 0
	}

	resp, err := (&http.Client{
		Timeout: time.Second * time.Duration(timeout),
	}).Do(req)
	if err != nil {
		fmt.Println("error performing request:", err)
		return 0
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response body:", err)
		return 0
	}

	L.Push(lua.LNumber(resp.StatusCode))
	L.Push(lua.LString(string(b)))
	return 2
}

func stringOrDefault(table *lua.LTable, key, defaultVal string) string {
	val := table.RawGetString(key)
	switch val.Type() {
	case lua.LTString:
		return val.String()
	default:
		return defaultVal
	}
}

func intOrDefault(table *lua.LTable, key string, defaultVal int) int {
	val := table.RawGetString(key)
	switch val.Type() {
	case lua.LTNumber:
		res, err := strconv.Atoi(val.String())
		if err != nil {
			return defaultVal
		}
		return res
	default:
		return defaultVal
	}
}
