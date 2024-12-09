package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
)

type HTTPServer struct {
	host string
	port string
}

func NewHTTPServer(host, port string) *HTTPServer {
	return &HTTPServer{
		host: host,
		port: port,
	}
}

func (h HTTPServer) Address() string {
	return net.JoinHostPort(h.host, h.port)
}

func (h *HTTPServer) startHTTPServer() error {
	http.HandleFunc("/send", h.sendHandler)

	log.Printf("HTTP сервер запущен на %s\n", h.Address())

	err := http.ListenAndServe(h.Address(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (h *HTTPServer) sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Ошибка при декодировании JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Получено сообщение через POST-запрос: Type=%s, Content=%s\n", msg.Type, msg.Content)

	err := processMessage(msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Сообщение успешно получено"))
}
