package exec

import (
	"encoding/json"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func marshalLValue(val lua.LValue) ([]byte, error) {
	switch val.Type() {
	case lua.LTNil:
		return json.Marshal(nil)
	case lua.LTBool:
		return json.Marshal(bool(val.(lua.LBool)))
	case lua.LTNumber:
		return json.Marshal(float64(val.(lua.LNumber)))
	case lua.LTString:
		return []byte(string(val.(lua.LString))), nil
	case lua.LTTable:
		t := val.(*lua.LTable)
		var outerErr error
		if lTableIsArray(t) {
			arr := make([]any, 0, t.Len())
			t.ForEach(func(_, l2 lua.LValue) {
				l2Val, err := lValueToGo(l2)
				if err != nil {
					outerErr = err
					return
				}
				arr = append(arr, l2Val)
			})
			if outerErr != nil {
				return nil, outerErr
			}
			return json.Marshal(arr)
		} else {
			m := make(map[string]any, t.Len())
			t.ForEach(func(l1, l2 lua.LValue) {
				l1Bytes, err := marshalLValue(l1)
				if err != nil {
					outerErr = err
					return
				}
				l2Val, err := lValueToGo(l2)
				if err != nil {
					outerErr = err
					return
				}
				m[string(l1Bytes)] = l2Val
			})
			return json.Marshal(m)
		}
	default:
		return nil, fmt.Errorf("unsupported value type: %s", val.Type().String())
	}
}

func lValueToGo(val lua.LValue) (any, error) {
	switch val.Type() {
	case lua.LTNil:
		return nil, nil
	case lua.LTBool:
		return bool(val.(lua.LBool)), nil
	case lua.LTNumber:
		return float64(val.(lua.LNumber)), nil
	case lua.LTString:
		return string(val.(lua.LString)), nil
	case lua.LTTable:
		t := val.(*lua.LTable)
		var outerErr error
		if lTableIsArray(t) {
			ret := make([]any, 0, t.Len())
			t.ForEach(func(_, l2 lua.LValue) {
				idxVal, err := lValueToGo(l2)
				if err != nil {
					outerErr = err
					return
				}
				ret = append(ret, idxVal)
			})
			return ret, outerErr
		} else {
			ret := make(map[string]any, t.Len())
			t.ForEach(func(l1, l2 lua.LValue) {
				l1Bytes, err := marshalLValue(l1)
				if err != nil {
					outerErr = err
					return
				}
				l2Val, err := lValueToGo(l2)
				if err != nil {
					outerErr = err
					return
				}
				ret[string(l1Bytes)] = l2Val
			})
			return ret, outerErr
		}
	default:
		return nil, fmt.Errorf("unsupported value type: %s", val.Type().String())
	}
}

func lTableIsArray(t *lua.LTable) bool {
	isArray := true
	idx := 0
	t.ForEach(func(l1, _ lua.LValue) {
		idx++
		l1Val, ok := l1.(lua.LNumber)
		if !ok {
			isArray = false
		} else if int(l1Val) != idx {
			isArray = false
		}
	})
	return isArray
}

func parseJSONString(payload json.RawMessage) (lua.LValue, error) {
	var err error

	var mapVal map[string]json.RawMessage
	if err = json.Unmarshal(payload, &mapVal); err == nil {
		ret := lua.LTable{}
		for k, v := range mapVal {
			lv, err := parseJSONString(v)
			if err != nil {
				return nil, err
			}
			ret.RawSetString(k, lv)
		}
		return &ret, nil
	}

	var arrVal []json.RawMessage
	if err = json.Unmarshal(payload, &arrVal); err == nil {
		ret := lua.LTable{}
		for i, v := range arrVal {
			lv, err := parseJSONString(v)
			if err != nil {
				return nil, err
			}
			ret.RawSetInt(i+1, lv)
		}
		return &ret, nil
	}

	var boolVal bool
	if err = json.Unmarshal(payload, &boolVal); err == nil {
		return lua.LBool(boolVal), nil
	}

	var floatVal float64
	if err = json.Unmarshal(payload, &floatVal); err == nil {
		return lua.LNumber(floatVal), nil
	}

	var stringVal string
	if err = json.Unmarshal(payload, &stringVal); err == nil {
		return lua.LString(stringVal), nil
	}

	return nil, fmt.Errorf("no suitable conversion found for JSON string '%s'", string(payload))
}
