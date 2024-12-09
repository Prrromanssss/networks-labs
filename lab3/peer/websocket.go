package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	upgrader websocket.Upgrader
	host     string
	port     string
}

func NewWebSocketServer(host, port string) *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		host: host,
		port: port,
	}
}

func (ws WebSocketServer) Address() string {
	return net.JoinHostPort(ws.host, ws.port)
}

func (ws *WebSocketServer) startWebSocketServer() error {
	http.HandleFunc("/ws", ws.handleConnection)

	log.Printf("WebSocket сервер запущен на %s\n", ws.Address())

	err := http.ListenAndServe(ws.Address(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (ws *WebSocketServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка при upgrading WebSocket:", err)
		return
	}
	defer conn.Close()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Ошибка при чтении сообщения из WebSocket:", err)
			break
		}

		if msg.Type == "get" {
			ws.sendCurrentAd(conn)
		}
	}
}

func (ws *WebSocketServer) sendCurrentAd(conn *websocket.Conn) {
	msg := Message{
		Type:    "get",
		Content: currentAd,
		From:    ws.Address(),
		ID:      currentAdID,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Println("Ошибка при отправке текущего объявления через WebSocket:", err)
	}
}
