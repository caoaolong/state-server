package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// //go:embed web/dist/**
// var distFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message:", err)
			return
		}
		log.Printf("Received message: %s", message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Failed to write message:", err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWebsocket)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
