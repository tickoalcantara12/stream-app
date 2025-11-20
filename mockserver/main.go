package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer conn.Close()

	// simple echo + broadcast to same connection
	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("server recv: %s\n", string(msg))

		// echo back
		if err := conn.WriteMessage(mt, msg); err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	log.Println("mock WS server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
