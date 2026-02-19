package main

import (
	"log"
	"net/http"

	"github.com/caoaolong/state-server/routers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// //go:embed web/dist/**
// var distFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebsocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
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
	r := gin.Default()
	r.GET("/ws", handleWebsocket)
	routers.RegisterStateMachineRoutes(r)
	routers.RegisterSessionRoutes(r)
	routers.RegisterApiKeyRoutes(r)
	routers.RegisterNodeRoutes(r)
	log.Fatal(r.Run(":8080"))
}
