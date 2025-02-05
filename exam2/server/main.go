package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("Square server listening on port 8081")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	fmt.Println("Received number:", msg)
	num, err := strconv.Atoi(msg[:len(msg)-1])
	if err != nil {
		fmt.Println("Conversion error:", err)
		return
	}

	result := num * num
	fmt.Fprintf(conn, "%d\n", result)
}
