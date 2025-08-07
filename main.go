package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all connections
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server started on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error: ", err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
			delete(clients, ws)
			break
		}
		type Message struct {
			Nickname  string `json:"nickname"`
			Message   string `json:"message"`
			Timestamp string `json:"timestamp"`
		}
		var message Message
		err = json.Unmarshal(msg, &message)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}
		message.Timestamp = time.Now().Format("02.01.2006 15:04:05")
		messageJSON, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshalling message: %v", err)
			continue
		}
		broadcast <- messageJSON
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			client.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
