package main

import (
	"fmt"
	"net"
	"sync"
)

func main() {
	address := net.JoinHostPort("0.0.0.0", "8080")

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		lis, err := net.Listen("tcp", address)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer lis.Close()

		fmt.Printf("TCP-сервер запущен на %s\n", address)

		for {
			conn, err := lis.Accept()
			if err != nil {
				fmt.Println("Connecting error:", err)
				continue
			}

			fmt.Println("Remote addr", conn.RemoteAddr())
		}
	}()

	go func() {
		defer wg.Done()

		upAddr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			fmt.Println(err)
			return
		}

		conn, err := net.ListenUDP("udp", upAddr)
		if err != nil {
			fmt.Println("Ошибка запуска сервера:", err)
			return
		}
		defer conn.Close()

		fmt.Printf("UDP-сервер запущен на %s\n", address)

		buffer := make([]byte, 1024)

		for {
			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Ошибка чтения данных:", err)
				continue
			}

			fmt.Printf("Получено сообщение: %s от %s\n", string(buffer[:n]), clientAddr)

			response := []byte("Сообщение получено")
			_, err = conn.WriteToUDP(response, clientAddr)
			if err != nil {
				fmt.Println("Ошибка отправки ответа:", err)
			}
		}
	}()

	fmt.Println("Waiting for servers")

	wg.Wait()
}
