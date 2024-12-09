package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

type TCPServer struct {
	host string
	port string
}

func NewTCPServer(host, port string) *TCPServer {
	return &TCPServer{
		host: host,
		port: port,
	}
}

func (t TCPServer) Address() string {
	return net.JoinHostPort(t.host, t.port)
}

func (t *TCPServer) startTCPServer() error {

	ln, err := net.Listen("tcp", t.Address())
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("TCP-сервер запущен на %s\n", t.Address())

	go t.checkUpdates()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Ошибка при подключении клиента:", err)
			continue
		}

		go t.handleConnection(conn)
	}
}

func (t *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		log.Println("Ошибка при декодировании сообщения:", err)
		return
	}

	if msg.From == currentAddr {
		return
	}

	switch msg.Type {
	case "post":
		if msg.ID != currentAdID || currentAdID == "" || msg.From == "CLIENT" {
			log.Printf("Получено новое объявление от %s: %s\n", msg.From, msg.Content)
			currentAd = msg.Content
			currentAdID = msg.ID
			forwardToParent(msg)
		}
	case "remove":
		log.Printf("Удаление объявления от %s\n", msg.From)
		currentAd = ""
		currentAdID = ""
		forwardToParent(msg)
	case "update":
		t.sendUpdate(conn)
	}
}

func (t *TCPServer) sendUpdate(conn net.Conn) {
	msg := Message{
		Type:    "post",
		Content: currentAd,
		From:    currentAddr,
		ID:      currentAdID,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(&msg); err != nil {
		log.Println("Ошибка при отправке обновления:", err)
	}
}

func (t *TCPServer) checkUpdates() {
	for {
		if parentAddr != "" {
			conn, err := net.Dial("tcp", parentAddr)
			if err != nil {
				log.Println("Не удалось подключиться к родителю для обновлений:", err)
				continue
			}

			msg := Message{
				Type: "update",
				From: currentAddr,
			}

			encoder := json.NewEncoder(conn)
			if err := encoder.Encode(&msg); err != nil {
				log.Println("Ошибка при отправке запроса на обновления:", err)
				conn.Close()
				continue
			}

			decoder := json.NewDecoder(conn)
			var receivedMsg Message
			if err := decoder.Decode(&receivedMsg); err == nil && receivedMsg.Type == "post" {
				if receivedMsg.ID != currentAdID || currentAdID == "" || msg.From == "CLIENT" {
					log.Printf("Получено обновление объявления: %s\n", receivedMsg.Content)
					currentAd = receivedMsg.Content
					currentAdID = receivedMsg.ID
				}
			}

			conn.Close()
		}
		time.Sleep(5 * time.Second)
	}
}
