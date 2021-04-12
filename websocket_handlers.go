package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// setupCORS(&w, r)
	// if (*r).Method == "OPTIONS" {
	// 	return
	// }

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connected")

	err = ws.WriteMessage(1, []byte("Yonkers motherfucker"))
	if err != nil {
		log.Println(err)
	}

	reader(ws)
}
