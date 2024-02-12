package exec

import (
	"fmt"
	"reflect"
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

func getStringParam(L *lua.LState, paramName string, stackPosition int) (string, error) {
	stackVal := L.Get(stackPosition)
	if stackVal.Type() != lua.LTString {
		return "", fmt.Errorf("error: getStringParam: expected '%s' parameter to be string, instead got '%s'", paramName, stackVal.Type().String())
	}
	return lua.LVAsString(stackVal), nil
}

func getStringArrayParam(L *lua.LState, paramName string, stackPosition int) ([]string, error) {
	stackVal := L.Get(stackPosition)
	if stackVal.Type() != lua.LTTable {
		return nil, fmt.Errorf("error: getStringArrayParam: expected '%s' parameter to be array, instead got '%s'", paramName, stackVal.Type().String())
	}
	arr := stackVal.(*lua.LTable)
	ret := make([]string, 0, arr.Len())
	arr.ForEach(func(_, l2 lua.LValue) {
		ret = append(ret, l2.String())
	})
	return ret, nil
}

func getMapParamWithDefault(L *lua.LState, paramName string, stackPosition int, def map[string]string) map[string]string {
	stackVal := L.Get(stackPosition)
	if stackVal.Type() != lua.LTTable {
		return def
	}
	arr := stackVal.(*lua.LTable)
	ret := make(map[string]string, arr.Len())
	arr.ForEach(func(l1, l2 lua.LValue) {
		ret[l1.String()] = l2.String()
	})
	return ret
}

func getIntParam(L *lua.LState, paramName string, stackPosition int) (int, error) {
	stackVal := L.Get(stackPosition)
	if val, ok := stackVal.(lua.LNumber); ok {
		return int(val), nil
	} else {
		return -1, fmt.Errorf("error: getIntParam: expected '%s' parameter to be integer, instead got '%s'", paramName, stackVal.Type().String())
	}
}

func luaTypeToString(val lua.LValue) (string, error) {
	switch val.Type() {
	case lua.LTString:
		return val.String(), nil
	default:
		return "", fmt.Errorf("incorrect type for value: %s\n", val.Type().String())
	}
}

func getString(table *lua.LTable, key string) (string, error) {
	val := table.RawGetString(key)
	switch val.Type() {
	case lua.LTString:
		return val.String(), nil
	default:
		return "", fmt.Errorf("expected type '%s', got: '%s'", lua.LTString.String(), val.Type().String())
	}
}

func stringOrDefault(table *lua.LTable, key, defaultVal string) string {
	res, err := getString(table, key)
	if err != nil {
		return defaultVal
	}
	return res
}

func getInt(table *lua.LTable, key string) (int, error) {
	val := table.RawGetString(key)
	switch val.Type() {
	case lua.LTNumber:
		res, err := strconv.Atoi(val.String())
		if err != nil {
			return -1, err
		}
		return res, nil
	default:
		return -1, fmt.Errorf("expected type '%s', got: '%s'", lua.LTNumber.String(), val.Type().String())
	}
}

func intOrDefault(table *lua.LTable, key string, defaultVal int) int {
	res, err := getInt(table, key)
	if err != nil {
		return defaultVal
	}
	return res
}

func getHeaderTable(table *lua.LTable, key string) (map[string][]string, error) {
	ret := make(map[string][]string)
	var err error

	innerTable := table.RawGetString(key)
	switch innerTable.Type() {
	case lua.LTTable:
		innerTable.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			var keyString string
			keyString, err = luaTypeToString(k)
			if err != nil {
				err = fmt.Errorf("error parsing header key '%s': %v", k, err)
				return
			}
			var valSlice []string
			valSlice, err = luaArrayToSlice(v)
			if err != nil {
				err = fmt.Errorf("error parsing header value '%s': %v", v, err)
				return
			}
			ret[keyString] = valSlice
		})
		if err != nil {
			return nil, err
		}
	case lua.LTNil:
		// this is fine, default to doing nothing
	default:
		return nil, fmt.Errorf("unexpected value found for header table slot. value: %v", innerTable.Type())
	}
	return ret, nil
}

func sliceToLuaArray(strs []string) *lua.LTable {
	arr := lua.LTable{}
	for _, s := range strs {
		arr.Append(lua.LString(s))
	}
	return &arr
}

func luaArrayToSlice(val lua.LValue) ([]string, error) {
	if val.Type() != lua.LTTable {
		return nil, fmt.Errorf("luaArrayToSlice expected table type, got '%s'", val.Type().String())
	}
	arr := val.(*lua.LTable)
	strs := make([]string, 0, arr.Len())
	arr.ForEach(func(_, l2 lua.LValue) {
		strs = append(strs, l2.String())
	})
	return strs, nil
}

func anyToLValue(arg any) (lua.LValue, error) {
	switch resolved := arg.(type) {
	case string:
		return lua.LString(resolved), nil
	case float64:
		return lua.LNumber(resolved), nil
	case int:
		return lua.LNumber(resolved), nil
	case bool:
		return lua.LBool(resolved), nil
	case nil:
		return lua.LNil, nil
	default:
		return lua.LNil, fmt.Errorf("no compatible Lua value found for '%v' of type '%v'", arg, reflect.TypeOf(arg))
	}
}
