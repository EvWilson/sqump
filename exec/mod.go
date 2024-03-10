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
	currentIdent Identifier
	currentEnv   string
	environment  map[string]string
	loopCheck    LoopChecker
	ctx          context.Context
	cancel       context.CancelFunc
	err          error
	oldReq       *lua.LFunction
}

type LoopChecker map[string]bool

func NewLoopChecker() LoopChecker {
	return make(LoopChecker)
}

func (lc LoopChecker) AddIdent(ident Identifier) bool {
	if _, ok := lc[ident.String()]; ok {
		return false
	}
	lc[ident.String()] = true
	return true
}

func (lc LoopChecker) ClearIdent(ident Identifier) {
	delete(lc, ident.String())
}

func CreateState(
	ident Identifier,
	currentEnv string,
	env data.EnvMapValue,
	loopCheck LoopChecker,
) *State {
	L := lua.NewState()
	ctx, cancel := context.WithCancel(context.Background())
	L.SetContext(ctx)

	state := State{
		LState:       L,
		currentIdent: ident,
		currentEnv:   currentEnv,
		environment:  env,
		loopCheck:    loopCheck,
		ctx:          ctx,
		cancel:       cancel,
		err:          nil,
		oldReq:       L.GetGlobal("require").(*lua.LFunction),
	}
	state.loopCheck.AddIdent(state.currentIdent)

	L.SetGlobal("print", L.NewFunction(printViaCore))
	L.SetGlobal("require", L.NewFunction(state.require))
	L.PreloadModule("sqump", func(l *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"fetch":          state.fetch,
			"print_response": state.printResponse,
			"to_json":        state.toJSON,
			"to_json_pretty": state.toJSONPretty,
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

func (s *State) toJSONPretty(_ *lua.LState) int {
	b, err := marshalLValue(s.LState.Get(1))
	if err != nil {
		return s.CancelErr("error: to_json_pretty: error serializing to JSON: %v", err)
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, "", "  ")
	if err != nil {
		return s.CancelErr("error: to_json_pretty: error indenting JSON: %v", err)
	}
	s.LState.Push(lua.LString(buf.String()))
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

func (s *State) require(_ *lua.LState) int {
	moduleName, err := getStringParam(s.LState, "module", 1)
	if err != nil {
		return s.CancelErr("error: require: %v", err)
	}
	coll, err := data.ReadCollection(s.currentIdent.Path)
	if err != nil {
		return s.CancelErr("error: require: %v", err)
	}
	for _, req := range coll.Requests {
		if moduleName == req.Name {
			ident := s.currentIdent
			ident.Request = req.Name
			if !s.loopCheck.AddIdent(ident) {
				return s.CancelErr("error: require: cyclical loop detected: '%s' calling '%s', which has already been executed. Loop checker state: %v", s.currentIdent.String(), ident.String(), s.loopCheck)
			}
			defer s.loopCheck.ClearIdent(ident)

			script, err := s.prepScript(req.Script.String())
			if err != nil {
				return s.CancelErr("error: require: %v", err)
			}
			err = s.DoString(script)
			if err != nil {
				return s.CancelErr("error: require: state error: %v, error: %v", s.err, err)
			}
			returned := s.LState.GetTop()
			for i := 1; i < returned; i++ {
				s.LState.Push(s.LState.Get(i + 1))
			}
			return returned - 1
		}
	}
	// Fall back to old require if needed
	s.LState.Push(s.oldReq)
	s.LState.Push(lua.LString(moduleName))
	s.LState.Call(1, 1)
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

func (s *State) prepScript(script string) (string, error) {
	return replaceEnvTemplates(s.currentIdent.String(), script, s.environment)
}
