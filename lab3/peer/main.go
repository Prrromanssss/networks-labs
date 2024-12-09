package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Использование: <ip> <tcp_port> <websocket_port> <http_port> <parent_tcp_ip:port (optional)>")
		return
	}

	host := os.Args[1]
	tcpPort := os.Args[2]
	webSocketPort := os.Args[3]
	httpPort := os.Args[4]
	var tcpParentAddr string
	if len(os.Args) > 5 {
		tcpParentAddr = os.Args[5]
	}

	httpServer := NewHTTPServer(host, httpPort)
	webSocketServer := NewWebSocketServer(host, webSocketPort)
	tcpServer := NewTCPServer(host, tcpPort)

	peer := NewPeer(host, tcpParentAddr, httpServer, webSocketServer, tcpServer)

	ctx, cancel := context.WithCancel(context.Background())

	peer.runPeer(ctx, cancel)
}
