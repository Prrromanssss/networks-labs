package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gofiber/fiber/v2/log"
)

type Peer struct {
	host          string
	parentAddress string
	tcp           *TCPServer
	http          *HTTPServer
	webSocket     *WebSocketServer
}

func NewPeer(
	host string,
	parentAddress string,
	http *HTTPServer,
	webSocket *WebSocketServer,
	tcp *TCPServer,
) *Peer {
	return &Peer{
		http:          http,
		webSocket:     webSocket,
		tcp:           tcp,
		host:          host,
		parentAddress: parentAddress,
	}
}

func (p *Peer) runPeer(ctx context.Context, cancel context.CancelFunc) {
	wg := &sync.WaitGroup{}

	wg.Add(3)

	go func() {
		defer wg.Done()

		err := p.webSocket.startWebSocketServer()
		if err != nil {
			log.Panicf("Error starting web socker server: %#v", err)
		}
	}()

	go func() {
		defer wg.Done()

		err := p.http.startHTTPServer()
		if err != nil {
			log.Panicf("Error starting http server: %#v", err)
		}
	}()

	go func() {
		defer wg.Done()

		err := p.tcp.startTCPServer()
		if err != nil {
			log.Panicf("Error starting tcp server: %#v", err)
		}
	}()

	p.gracefulShutdown(ctx, cancel, wg)
}

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	From    string `json:"from"`
	ID      string `json:"id"`
}

var parentAddr string
var currentAddr string
var currentAd string
var currentAdID string

func processMessage(msg Message) error {
	switch msg.Type {
	case "post":
		postAd(currentAddr, msg.Content, msg.ID)
	case "remove":
		removeAd(currentAddr)
	case "print":
		printAd()
	default:
		return errors.New("unknown command")
	}

	return nil
}

func postAd(addr, ad, id string) {
	msg := Message{
		Type:    "post",
		Content: ad,
		From:    addr,
		ID:      id,
	}

	if parentAddr == "" {
		log.Info("Это корневой узел, сообщение остается здесь")
		currentAd = ad
		currentAdID = id
	} else {
		forwardToParent(msg)
	}
}

func forwardToParent(msg Message) {
	if parentAddr == "" {
		log.Info("Это корневой узел, пересылка не требуется")
		return
	}
	conn, err := net.Dial("tcp", parentAddr)
	if err != nil {
		log.Error("Не удалось подключиться к родителю:", err)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		log.Error("Ошибка при отправке сообщения родителю:", err)
	}
}

func removeAd(addr string) {
	msg := Message{
		Type: "remove",
		From: addr,
	}

	if parentAddr == "" {
		log.Info("Это корневой узел, удаляем здесь")
		currentAd = ""
		currentAdID = ""
	} else {
		forwardToParent(msg)
	}
}

func printAd() {
	if currentAd != "" {
		log.Infof("Текущее объявление: %s\n", currentAd)
	} else {
		log.Info("[Объявление отсутствует]")
	}
}

func (p *Peer) gracefulShutdown(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	select {
	case <-ctx.Done():
		log.Info("terminating: context cancelled")
	case <-waitSignal():
		log.Info("terminating: via signal")
	}

	cancel()
	if wg != nil {
		wg.Wait()
	}
}

func waitSignal() chan os.Signal {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	return sigterm
}
