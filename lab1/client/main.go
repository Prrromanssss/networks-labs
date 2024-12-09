package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"lab1/models"
	"os"
	"strconv"
	"strings"

	"log"
	"net"
)

const (
	baseHost = "185.102.139.168"
	basePort = "50051"
)

func main() {
	address := net.JoinHostPort(baseHost, basePort)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Cannot connect to server:", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Type numbers:")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		strNumbers := strings.Split(input, " ")

		var numbers []int
		for _, strNum := range strNumbers {
			num, err := strconv.Atoi(strNum)
			if err != nil {
				log.Fatal("Wrong format:", err)
			}
			numbers = append(numbers, num)
		}

		fmt.Println("Type sort (asc or desc):")
		order, _ := reader.ReadString('\n')
		order = strings.TrimSpace(order)

		if order != "asc" && order != "desc" {
			log.Fatal("Wrong type of sort. Waits for 'asc' or 'desc'.")
		}

		req := models.SortRequest{
			Array: numbers,
			Order: order,
		}

		err = encoder.Encode(&req)
		if err != nil {
			log.Fatal("Error in sending request:", err)
		}

		var res models.SortResponse
		err = decoder.Decode(&res)
		if err != nil {
			var errMsg models.ErrorMessage
			err = decoder.Decode(&errMsg)
			if err != nil {
				log.Fatal("Connection error:", err)
			}
			log.Println("Error:", errMsg.Message)
			return
		}

		log.Printf("Init array: %v\n", res.OriginalArray)
		log.Printf("Sorted array: %v\n", res.SortedArray)
	}
}
