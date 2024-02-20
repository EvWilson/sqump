package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/prnt"

	lua "github.com/yuin/gopher-lua"
)

type State struct {
	*lua.LState
	config       *data.Config
	exports      ExportMap
	currentIdent Identifier
	environment  map[string]string
	loopCheck    LoopChecker
	ctx          context.Context
	cancel       context.CancelFunc
	err          error
}

type LoopChecker map[string]bool

type ExportMap map[string]*lua.LTable

func CreateState(
	conf *data.Config,
	ident Identifier,
	env data.EnvMapValue,
	loopCheck LoopChecker,
) *State {
	L := lua.NewState()
	ctx, cancel := context.WithCancel(context.Background())
	L.SetContext(ctx)

	state := State{
		LState:       L,
		config:       conf,
		exports:      make(ExportMap),
		currentIdent: ident,
		environment:  env,
		loopCheck:    loopCheck,
		ctx:          ctx,
		cancel:       cancel,
		err:          nil,
	}
	state.loopCheck[state.currentIdent.String()] = true

	L.SetGlobal("print", L.NewFunction(printViaCore))
	L.PreloadModule("sqump", func(l *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"execute":        state.execute,
			"export":         state.export,
			"fetch":          state.fetch,
			"print_response": state.printResponse,
			"to_json":        state.toJSON,
			"from_json":      state.fromJSON,
		})
		L.Push(mod)
		return 1
	})
	state.registerKafkaModule(L)

	return &state
}

// https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API
func (s *State) fetch(_ *lua.LState) int {
	// Get params
	resource, err := getStringParam(s.LState, "resource", 1)
	if err != nil {
		return s.CancelErr("error: fetch: %v", err)
	}

	optionVal := s.LState.Get(2)
	var options *lua.LTable
	switch optionVal.Type() {
	case lua.LTTable:
		options = optionVal.(*lua.LTable)
	case lua.LTNil:
		options = &lua.LTable{}
	default:
		return s.CancelErr("error: fetch: expected 'options' parameter to be table or nil, instead got: %s", optionVal.Type().String())
	}

	// Marshal body
	var buf *bytes.Buffer
	body := options.RawGetString("body")
	switch body.Type() {
	case lua.LTTable:
		bodyMap := make(map[string]any)
		body.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			keyString, err := luaTypeToString(k)
			if err != nil {
				_ = s.CancelErr("error: fetch: while parsing body key '%v': %v", k, err)
				return
			}
			bodyMap[keyString] = v
		})
		b, err := json.Marshal(bodyMap)
		if err != nil {
			return s.CancelErr("error: fetch: while marshaling body table in fetch: %v", err)
		}
		buf = bytes.NewBuffer(b)
	case lua.LTString:
		str, err := luaTypeToString(body)
		if err != nil {
			return s.CancelErr("error: fetch: while converting body string in fetch: %v", err)
		}
		buf = bytes.NewBuffer([]byte(str))
	case lua.LTNil:
		buf = bytes.NewBuffer([]byte{})
	default:
		return s.CancelErr("error: fetch: unsupported body type: %s. expected: table, string, or nil", body.Type().String())
	}

	// Get other option items
	method := stringOrDefault(options, "method", "GET")
	timeout := intOrDefault(options, "timeout", 10)

	req, err := http.NewRequest(method, resource, buf)
	if err != nil {
		return s.CancelErr("error: fetch: while creating request: %v", err)
	}

	// Add headers
	req.Header.Add("User-Agent", "sqump")
	reqHeaderTable := options.RawGetString("headers")
	switch reqHeaderTable.Type() {
	case lua.LTTable:
		reqHeaderTable.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			keyString, err := luaTypeToString(k)
			if err != nil {
				_ = s.CancelErr("error: fetch: while parsing header key '%s': %v", k, err)
				return
			}
			valString, err := luaTypeToString(v)
			if err != nil {
				_ = s.CancelErr("error: fetch: while parsing header value '%s': %v", v, err)
				return
			}
			req.Header.Add(keyString, valString)
		})
	case lua.LTNil:
		// this is fine, default to doing nothing
	default:
		return s.CancelErr("error: fetch: unexpected value found for header table slot. value: %v", reqHeaderTable.Type())
	}

	// Perform request
	resp, err := (&http.Client{
		Timeout: time.Second * time.Duration(timeout),
	}).Do(req)
	if err != nil {
		return s.CancelErr("error: fetch: while performing request: %v", err)
	}
	defer resp.Body.Close()

	// Gather headers into a lua table
	respHeaderTable := &lua.LTable{}
	for k, v := range resp.Header {
		respHeaderTable.RawSetString(k, sliceToLuaArray(v))
	}

	// Read response
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return s.CancelErr("error: fetch: while reading response body: %v", err)
	}

	respTable := &lua.LTable{}
	respTable.RawSetString("status", lua.LNumber(resp.StatusCode))
	respTable.RawSetString("headers", respHeaderTable)
	respTable.RawSetString("body", lua.LString(string(b)))
	s.LState.Push(respTable)
	return 1
}

func (s *State) execute(_ *lua.LState) int {
	request, err := getStringParam(s.LState, "request", 1)
	if err != nil {
		return s.CancelErr("error: execute: %v", err)
	}
	overrides, err := getMapParamOrNil(s.LState, "overrides", 2)
	if err != nil {
		return s.CancelErr("error: execute: %v", err)
	}

	ident := s.currentIdent
	ident.Request = request

	if _, ok := s.loopCheck[ident.String()]; ok {
		return s.CancelErr("error: execute: cyclical loop detected: '%s' calling '%s', which has already been executed. Loop checker state: %v", s.currentIdent.String(), ident.String(), s.loopCheck)
	}

	coll, err := data.ReadCollection(ident.Path)
	if err != nil {
		return s.CancelErr("error: execute: while reading collection at '%s': %v", ident.Path, err)
	}
	state, err := ExecuteRequest(coll, request, s.config, mergeMaps(s.environment, overrides), s.loopCheck)
	if err != nil {
		return s.CancelErr("error: execute: while performing request '%s': %v", request, err)
	}

	export, ok := state.exports[ident.String()]
	if ok {
		s.exports[ident.String()] = export
	}

	s.LState.Push(export)
	return 1
}

func (s *State) export(_ *lua.LState) int {
	ctxVal := s.LState.Get(1)
	if ctxVal.Type() != lua.LTTable {
		return s.CancelErr("error: export: expected context parameter to be table, instead got: %s", ctxVal.Type().String())
	}
	ctx := ctxVal.(*lua.LTable)
	s.exports[s.currentIdent.String()] = ctx

	return 0
}

func (s *State) printResponse(_ *lua.LState) int {
	respVal := s.LState.Get(1)
	if respVal.Type() != lua.LTTable {
		return s.CancelErr("error: print_response: expected response parameter to be table, instead got: %s", respVal.Type().String())
	}
	resp := respVal.(*lua.LTable)

	code, err := getInt(resp, "status")
	if err != nil {
		return s.CancelErr("error: print_response: while retrieving status code: %v", err)
	}
	headers, err := getHeaderTable(resp, "headers")
	if err != nil {
		return s.CancelErr("error: print_response: while retrieving map from table: %v", err)
	}
	body, err := getString(resp, "body")
	if err != nil {
		return s.CancelErr("error: print_response: while retrieving body: %v", err)
	}

	isJson := func(headers map[string][]string) bool {
		for k, v := range headers {
			if k == "Content-Type" {
				for _, header := range v {
					if strings.Contains(header, "application/json") {
						return true
					}
				}
				break
			}
		}
		return false
	}
	if isJson(headers) {
		var buf bytes.Buffer
		err := json.Indent(&buf, []byte(body), "", "  ")
		if err != nil {
			return s.CancelErr("error: print_response: while trying to indent json response body: %v", err)
		}
		body = buf.String()
	}

	prnt.Printf("Status Code: %d\n\n", code)
	prnt.Println("Headers:")
	for k, v := range headers {
		prnt.Printf("%s: %s\n", k, v)
	}
	prnt.Printf("\nBody:\n%s\n", body)

	return 0
}

func (s *State) toJSON(_ *lua.LState) int {
	b, err := marshalLValue(s.LState.Get(1))
	if err != nil {
		return s.CancelErr("error: to_json: error serializing to JSON: %v", err)
	}
	s.LState.Push(lua.LString(b))
	return 1
}

func (s *State) fromJSON(_ *lua.LState) int {
	payload, err := getStringParam(s.LState, "json", 1)
	if err != nil {
		return s.CancelErr("error: from_json: %v", err)
	}
	lv, err := parseJSONString([]byte(payload))
	if err != nil {
		return s.CancelErr("error: from_json: error parsing JSON string: %v", err)
	}
	s.LState.Push(lv)
	return 1
}

func printViaCore(L *lua.LState) int {
	top := L.GetTop()
	args := make([]interface{}, 0, top)
	for i := 1; i <= top; i++ {
		args = append(args, L.ToStringMeta(L.Get(i)).String())
	}
	prnt.Println(args...)
	return 0
}

func (s *State) CancelErr(format string, args ...any) int {
	s.err = fmt.Errorf(format, args...)
	s.cancel()
	return 0
}
