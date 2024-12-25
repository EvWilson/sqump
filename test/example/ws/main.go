package main

import (
	"fmt"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	fmt.Println("WS echo server listening at localhost:8080")
	_ = http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			panic(err)
		}
		go func() {
			defer conn.Close()
			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Got msg, echoing: %s\n", msg)
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					panic(err)
				}
			}
		}()
	}))
}
