package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	for {
		conn, err := net.Dial("tcp", "localhost:8081")
		if err != nil {
			panic(err)
		}

		fmt.Print("Enter a number: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		fmt.Fprintf(conn, "%s\n", input)

		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			panic(err)
		}

		fmt.Println("Squared result:", response)
		conn.Close()
	}
}
