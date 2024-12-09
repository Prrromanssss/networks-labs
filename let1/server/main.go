package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка обновления до WebSocket:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Ошибка чтения:", err)
			break
		}

		customMessage := map[string]string{
			"type":                "response",
			"content":             "Message received",
			"message from client": string(message),
		}
		msg, err := json.Marshal(customMessage)
		if err != nil {
			log.Println("Ошибка маршалинга:", err)
			break
		}

		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Ошибка записи:", err)
			break
		}
	}

	log.Println("Соединение закрыто нормально")
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	log.Println("Server started ")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
