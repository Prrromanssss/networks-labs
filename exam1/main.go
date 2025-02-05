package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Str1 string `json:"str1"`
	Str2 string `json:"str2"`
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Ошибка подключения:", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Ошибка чтения:", err)
			break
		}

		var data Message
		if err := json.Unmarshal(msg, &data); err != nil {
			fmt.Println("Ошибка парсинга JSON:", err)
			continue
		}

		result := fmt.Sprintf("%s-%s", data.Str1, data.Str2)

		conn.WriteMessage(websocket.TextMessage, []byte(result))
	}
}

func main() {
	http.HandleFunc("/ws", handleConnection)
	fmt.Println("Сервер запущен на порту 8080")
	http.ListenAndServe(":8080", nil)
}
