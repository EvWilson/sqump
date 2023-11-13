package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type State struct {
	*lua.LState
	exports      ExportMap
	currentIdent Identifier
	environment  map[string]string
}

type ExportMap map[string]*lua.LTable

func CreateState(ident Identifier, env map[string]string) *State {
	L := lua.NewState()

	state := State{
		LState:       L,
		exports:      make(ExportMap),
		currentIdent: ident,
		environment:  env,
	}

	L.PreloadModule("sqump", func(l *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"execute": state.execute,
			"export":  state.export,
			"fetch":   state.fetch,
		})
		L.Push(mod)
		return 1
	})

	return &state
}

// https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API
func (s *State) fetch(L *lua.LState) int {
	// Get params
	resourceVal := L.Get(1)
	if resourceVal.Type() != lua.LTString {
		fmt.Printf("expected resource parameter to be string, instead got: %s\n", resourceVal.Type().String())
		os.Exit(1)
	}
	resource := lua.LVAsString(resourceVal)
	optionVal := L.Get(2)
	if optionVal.Type() != lua.LTTable {
		fmt.Printf("expected options parameter to be table, instead got: %s\n", optionVal.Type().String())
		os.Exit(1)
	}
	options := optionVal.(*lua.LTable)

	// Marshal body
	var buf *bytes.Buffer
	body := options.RawGetString("body")
	switch body.Type() {
	case lua.LTTable:
		bodyMap := make(map[string]any)
		body.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			keyString, err := luaTypeToString(k)
			if err != nil {
				fmt.Printf("error parsing body key '%v': %v\n", k, err)
				os.Exit(1)
			}
			bodyMap[keyString] = v
		})
		b, err := json.Marshal(bodyMap)
		if err != nil {
			fmt.Println("error marshaling body table in fetch:", err)
			os.Exit(1)
		}
		buf = bytes.NewBuffer(b)
	case lua.LTString:
		str, err := luaTypeToString(body)
		if err != nil {
			fmt.Println("error converting body string in fetch:", err)
			os.Exit(1)
		}
		buf = bytes.NewBuffer([]byte(str))
	case lua.LTNil:
		buf = bytes.NewBuffer([]byte{})
	default:
		fmt.Printf("unsupported body type for fetch: %s\n", body.Type().String())
		fmt.Println("expected: table, string, or nil")
		os.Exit(1)
	}

	// Get other option items
	method := stringOrDefault(options, "method", "GET")
	timeout := intOrDefault(options, "timeout", 10)

	req, err := http.NewRequest(method, resource, buf)
	if err != nil {
		fmt.Println("error creating request:", err)
		os.Exit(1)
	}

	// Add headers
	headerTable := options.RawGetString("headers")
	switch headerTable.Type() {
	case lua.LTTable:
		headerTable.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			keyString, err := luaTypeToString(k)
			if err != nil {
				fmt.Println("error parsing header key:", err)
				os.Exit(1)
			}
			valString, err := luaTypeToString(v)
			if err != nil {
				fmt.Println("error parsing header value:", err)
				os.Exit(1)
			}
			req.Header.Add(keyString, valString)
		})
	case lua.LTNil:
		// this is fine, default to doing nothing
	default:
		fmt.Printf("unexpected value found for header table slot. value: %v\n", headerTable.Type())
		os.Exit(1)
	}

	// Perform request
	resp, err := (&http.Client{
		Timeout: time.Second * time.Duration(timeout),
	}).Do(req)
	if err != nil {
		fmt.Println("error performing request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read and push out response
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response body:", err)
		os.Exit(1)
	}

	L.Push(lua.LNumber(resp.StatusCode))
	L.Push(lua.LString(string(b)))
	return 2
}

func (s *State) execute(L *lua.LState) int {
	requestVal := L.Get(1)
	if requestVal.Type() != lua.LTString {
		fmt.Printf("error: expected request parameter to be string, instead got: %s\n", requestVal.Type().String())
		os.Exit(1)
	}
	request := lua.LVAsString(requestVal)

	if request == s.currentIdent.Request {
		fmt.Printf("error: self-referential request execution detected for '%s'\n", request)
		os.Exit(1)
	}
	ident := s.currentIdent
	ident.Request = request

	sq, err := ReadSqumpfile(ident.Path)
	if err != nil {
		fmt.Printf("error reading squmpfile at '%s': %v\n", ident.Path, err)
		os.Exit(1)
	}
	state, err := sq.ExecuteRequest(request)
	if err != nil {
		fmt.Printf("error performing request '%s': %v\n", request, err)
		os.Exit(1)
	}

	export, ok := state.exports[ident.String()]
	if ok {
		s.exports[ident.String()] = export
	}

	L.Push(export)
	return 1
}

func (s *State) export(L *lua.LState) int {
	ctxVal := L.Get(1)
	if ctxVal.Type() != lua.LTTable {
		fmt.Printf("expected context parameter to be table, instead got: %s\n", ctxVal.Type().String())
		os.Exit(1)
	}
	ctx := ctxVal.(*lua.LTable)
	s.exports[s.currentIdent.String()] = ctx

	return 0
}

func luaTypeToString(val lua.LValue) (string, error) {
	switch val.Type() {
	case lua.LTString:
		return val.String(), nil
	default:
		return "", fmt.Errorf("incorrect type for value: %s\n", val.Type().String())
	}
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
