package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type UnparsedCommand struct {
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload"`
}

type Command struct {
	Name    string `json:"command"`
	Payload any    `json:"payload"`
}

type ViewRequestPayload struct {
	EscapedPath string `json:"path"`
	Title       string `json:"title"`
	Scope       string `json:"scope"`
	Environment string `json:"environment"`
}

type ViewResponsePayload struct {
	ReplacedScript string `json:"script"`
}

type ExecRequestPayload struct {
	EscapedPath string `json:"path"`
	Title       string `json:"title"`
	Scope       string `json:"scope"`
	Environment string `json:"environment"`
}

type ExecResponsePayload struct {
	OutputFragment string `json:"fragment"`
}

func (r *Router) handleSocketConnection(w http.ResponseWriter, req *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(req, w)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.l.Debug("ws connection opened")
	core.SetPrinter(core.NewDualWriter(
		func(msg string, args ...any) (int, error) {
			formatted := fmt.Sprintf(msg, args...)
			cmd := Command{
				Name: "exec",
				Payload: ExecResponsePayload{
					OutputFragment: formatted,
				},
			}
			b, err := json.Marshal(cmd)
			if err != nil {
				return 0, err
			}
			err = wsutil.WriteServerMessage(conn, ws.OpText, b)
			if err != nil {
				r.l.Error("error writing server message from Printf", "error", err)
			}
			return -1, err
		},
		func(args ...any) (int, error) {
			formatted := fmt.Sprintln(args...)
			cmd := Command{
				Name: "exec",
				Payload: ExecResponsePayload{
					OutputFragment: formatted,
				},
			}
			b, err := json.Marshal(cmd)
			if err != nil {
				return 0, err
			}
			err = wsutil.WriteServerMessage(conn, ws.OpText, b)
			if err != nil {
				r.l.Error("error writing server message from Println", "error", err)
			}
			return -1, err
		},
	))
	go func() {
		defer func() {
			core.SetPrinter(&core.StandardPrinter{})
			err = conn.Close()
			if err != nil {
				r.l.Error("error closing connection", "error", err)
				return
			}
		}()

		for {
			msg, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				var ce wsutil.ClosedError
				if errors.As(err, &ce) {
					if ce.Code == ws.StatusNoStatusRcvd || ce.Code == ws.StatusGoingAway {
						return
					}
				}
				r.ServerError(w, err)
				return
			}
			var cmd UnparsedCommand
			err = json.Unmarshal(msg, &cmd)
			if err != nil {
				r.ServerError(w, err)
				return
			}
			r.l.Debug("received message", "command", cmd)
			switch cmd.Command {
			case "view":
				err = handleViewCommand(conn, cmd.Payload)
				if err != nil {
					core.Println("error encountered in view command:", err)
				}
			case "exec":
				err = handleExecCommand(conn, cmd.Payload)
				if err != nil {
					core.Println("error encountered in exec command:", err)
				}
			default:
				r.ServerError(w, fmt.Errorf("unrecognized command: %s\n", cmd.Command))
				return
			}
		}
	}()
}

func handleViewCommand(conn net.Conn, payload json.RawMessage) error {
	var data ViewRequestPayload
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return err
	}
	path, err := url.PathUnescape(data.EscapedPath)
	if err != nil {
		return err
	}
	var overrides core.EnvMapValue
	if data.Scope == "temp" {
		var ok bool
		overrides, ok = getTempConfig()[data.Environment]
		if !ok {
			return fmt.Errorf("no overrides found for environment '%s'", data.Environment)
		}
	}
	prepared, err := handlers.GetPreparedScript("/"+path, data.Title, overrides)
	if err != nil {
		return err
	}
	b, err := json.Marshal(Command{
		Name: "replaced",
		Payload: ViewResponsePayload{
			ReplacedScript: prepared,
		},
	})
	if err != nil {
		return err
	}
	return wsutil.WriteServerMessage(conn, ws.OpText, b)
}

func handleExecCommand(conn net.Conn, payload json.RawMessage) error {
	err := sendClearCommand(conn)
	if err != nil {
		return err
	}
	var data ExecRequestPayload
	err = json.Unmarshal(payload, &data)
	if err != nil {
		return err
	}
	path, err := url.PathUnescape(data.EscapedPath)
	if err != nil {
		return err
	}
	var overrides core.EnvMapValue
	if data.Scope == "temp" {
		var ok bool
		overrides, ok = getTempConfig()[data.Environment]
		if !ok {
			return fmt.Errorf("no overrides found for environment '%s'", data.Environment)
		}
	}
	err = handlers.ExecuteRequest("/"+path, data.Title, overrides)
	if err != nil {
		return err
	}
	return nil
}

func sendClearCommand(conn net.Conn) error {
	return wsutil.WriteServerMessage(conn, ws.OpText, []byte("{\"command\":\"clear\"}"))
}
