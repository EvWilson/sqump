package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/prnt"

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
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	Environment string `json:"environment"`
}

type ViewResponsePayload struct {
	ReplacedScript string `json:"script"`
}

type ExecRequestPayload struct {
	EscapedPath string `json:"path"`
	Name        string `json:"name"`
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
	prnt.SetPrinter(prnt.NewDualWriter(
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
			prnt.SetPrinter(&prnt.StandardPrinter{})
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
					prnt.Println("error encountered in view command:", err)
				}
			case "exec":
				err = handleExecCommand(conn, cmd.Payload)
				if err != nil {
					prnt.Println("error encountered in exec command:", err)
				}
			default:
				r.ServerError(w, fmt.Errorf("unrecognized command: %s\n", cmd.Command))
				return
			}
		}
	}()
}

func handleViewCommand(conn net.Conn, payload json.RawMessage) error {
	var vrp ViewRequestPayload
	err := json.Unmarshal(payload, &vrp)
	if err != nil {
		return err
	}
	path, err := url.PathUnescape(vrp.EscapedPath)
	if err != nil {
		return err
	}
	var overrides data.EnvMapValue
	if vrp.Scope == "temp" {
		var ok bool
		overrides, ok = getTempConfig()[vrp.Environment]
		if !ok {
			return fmt.Errorf("no overrides found for environment '%s'", vrp.Environment)
		}
	}
	prepared, err := handlers.GetPreparedScript(fmt.Sprintf("/%s", path), vrp.Name, overrides)
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
	var erp ExecRequestPayload
	err = json.Unmarshal(payload, &erp)
	if err != nil {
		return err
	}
	path, err := url.PathUnescape(erp.EscapedPath)
	if err != nil {
		return err
	}
	var overrides data.EnvMapValue
	if erp.Scope == "temp" {
		var ok bool
		overrides, ok = getTempConfig()[erp.Environment]
		if !ok {
			return fmt.Errorf("no overrides found for environment '%s'", erp.Environment)
		}
	}
	err = handlers.ExecuteRequest(fmt.Sprintf("/%s", path), erp.Name, overrides)
	if err != nil {
		return err
	}
	return nil
}

func sendClearCommand(conn net.Conn) error {
	return wsutil.WriteServerMessage(conn, ws.OpText, []byte("{\"command\":\"clear\"}"))
}
