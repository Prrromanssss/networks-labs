package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	From    string `json:"from"`
	ID      string `json:"id"`
}

var ads []string
var mu sync.Mutex

func fetchAds(wsURLs []string) {
	mu.Lock()
	defer mu.Unlock()

	ads = nil

	for _, wsURL := range wsURLs {
		u := url.URL{Scheme: "ws", Host: wsURL, Path: "/ws"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("Ошибка подключения к %s: %v\n", wsURL, err)
			continue
		}
		defer conn.Close()

		err = conn.WriteJSON(Message{Type: "print"})
		if err != nil {
			log.Printf("Ошибка отправки сообщения на %s: %v\n", wsURL, err)
			continue
		}

		var msg Message
		err = conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Ошибка при чтении ответа от %s: %v\n", wsURL, err)
			continue
		}

		if msg.Type == "print" {
			ads = append(ads, msg.Content)
		}
	}
}

func handleSendReq(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST-запросы поддерживаются", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Ошибка при парсинге формы", http.StatusBadRequest)
		return
	}

	msg := Message{
		Type:    r.FormValue("type"),
		Content: r.FormValue("content"),
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		From:    "CLIENT",
	}

	log.Printf("Получено сообщение: Type=%s, Content=%s\n", msg.Type, msg.Content)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Сообщение получено: %s", msg.Content)

	targetURL := "http://185.102.139.161:50055/send"

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Ошибка при преобразовании сообщения в JSON: %v\n", err)
		return
	}

	resp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		log.Printf("Ошибка при отправке сообщения на %s: %v\n", targetURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Неуспешный ответ от сервера: %s\n", resp.Status)
		return
	}

	log.Println("Сообщение успешно отправлено на другой сервер.")

}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintln(w, "<html><body>")
	fmt.Fprintln(w, "<h1>Список объявлений</h1>")
	fmt.Fprintln(w, "<ul>")
	for i, ad := range ads {
		fmt.Fprintf(w, "<li>Peer: %d %s</li>\n", i, ad)
	}
	fmt.Fprintln(w, "</ul>")
	fmt.Fprintln(w, "</body></html>")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: <ws_url_1> <ws_url_2> <ws_url_3> <ws_url_4>")
		return
	}

	wsURLs := os.Args[1:]

	http.HandleFunc("/", handleHTTP)
	http.HandleFunc("/send", handleSendReq)

	go func() {
		for {
			fetchAds(wsURLs)

		}
	}()

	fmt.Println("Запуск сервера на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
