package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func readIncomingWsMsg(conn *websocket.Conn) {
	for {
		// messageType, p, err := conn.ReadMessage()
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(p))
		// fmt.Printf("Msg type: %v\n", messageType)
	}
}

func wsConnectHandler(w http.ResponseWriter, r *http.Request) {
	// setupCORS(&w, r)
	// if (*r).Method == "OPTIONS" {
	// 	return
	// }

	ws, _ := upgrader.Upgrade(w, r, nil)
	log.Println("Client Connected")

	//save connection globally
	m := make(map[string]*websocket.Conn)
	m[mux.Vars(r)["id"]] = ws
	wsConnections = append(wsConnections, m)

	for _, mymap := range wsConnections {
		keys := make([]string, 0, len(mymap))
		for k, _ := range mymap {
			keys = append(keys, k)
		}
		fmt.Println(keys)
	}

	err := ws.WriteMessage(1, []byte("Yonkers motherfucker"))
	if err != nil {
		log.Println(err)
	}

	go readIncomingWsMsg(ws)
}
