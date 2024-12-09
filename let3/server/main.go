package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Сервер запущен на :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка при принятии подключения:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		message := time.Now().Format("15:04:05") + " GOOOOOL!!!\n"

		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Ошибка при отправке данных:", err)
			return
		}

		time.Sleep(time.Second)
	}
}
