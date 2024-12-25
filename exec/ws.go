package exec

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaWebsocketClientTypeName = "wsclient"
)

type WSClient struct {
	conn net.Conn
	url  string
}

func (ws *WSClient) toUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = ws
	L.SetMetatable(ud, L.GetTypeMetatable(luaWebsocketClientTypeName))
	return ud
}

func getClientParam(L *lua.LState, i int) (*WSClient, error) {
	v := L.Get(i)
	ud, ok := v.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("error: getClientParam: expected user data type for 'wsclient', got: '%s'", v.Type().String())
	}
	if v, ok := ud.Value.(*WSClient); ok {
		return v, nil
	}
	return nil, fmt.Errorf("error: getClientParam: expected 'WSClient' for 'wsclient', got: '%s'", reflect.TypeOf(ud.Value).String())
}

func (s *State) registerWebsocketModule(L *lua.LState) {
	L.PreloadModule("sqump_ws", func(l *lua.LState) int {
		// Register client type
		{
			clientMT := L.NewTypeMetatable(luaWebsocketClientTypeName)
			L.SetGlobal(luaWebsocketClientTypeName, clientMT)
			L.SetField(clientMT, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
				"send":      s.sendMessage,
				"onmessage": s.onMessage,
				"close":     s.close,
			}))
		}

		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"new_client": s.newClient,
		})
		L.Push(mod)
		return 1
	})
}

func (s *State) newClient(_ *lua.LState) int {
	urlStr, err := getStringParam(s.LState, "url", 1)
	if err != nil {
		return s.CancelErr("error: new_client: %v", err)
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return s.CancelErr("error: new_client: %v", err)
	}
	conn, _, _, err := ws.Dial(context.Background(), u.String())
	if err != nil {
		return s.CancelErr("error: new_client: %v", err)
	}
	c := WSClient{
		conn: conn,
		url:  urlStr,
	}
	s.LState.Push(c.toUserData(s.LState))
	return 1
}

func (s *State) sendMessage(_ *lua.LState) int {
	client, err := getClientParam(s.LState, 1)
	if err != nil {
		return s.CancelErr("error: send_message: %v", err)
	}
	msg, err := getStringParam(s.LState, "msg", 2)
	if err != nil {
		return s.CancelErr("error: send_message: %v", err)
	}
	err = wsutil.WriteClientText(client.conn, []byte(msg))
	if err != nil {
		return s.CancelErr("error: send_message: %v", err)
	}
	return 0
}

func (s *State) onMessage(_ *lua.LState) int {
	client, err := getClientParam(s.LState, 1)
	if err != nil {
		return s.CancelErr("error: onmessage: %v", err)
	}
	cb, err := getFuncParam(s.LState, "cb", 2)
	if err != nil {
		return s.CancelErr("error: onmessage: %v", err)
	}

	go func() {
		for {
			msg, err := wsutil.ReadServerText(client.conn)
			if err != nil {
				var errType *net.OpError
				if errors.As(err, &errType) {
					break
				}
				s.CancelErr("error: onmessage: %v", err)
			}
			s.LState.Push(cb)
			s.LState.Push(lua.LString(string(msg)))
			if err := s.LState.PCall(1, 0, nil); err != nil {
				s.CancelErr("error: onmessage: %v", err)
			}
		}
	}()

	return 0
}

func (s *State) close(_ *lua.LState) int {
	client, err := getClientParam(s.LState, 1)
	if err != nil {
		return s.CancelErr("error: close: %v", err)
	}
	err = client.conn.Close()
	if err != nil {
		return s.CancelErr("error: close: %v", err)
	}
	return 0
}
