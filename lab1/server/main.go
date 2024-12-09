package main

import (
	"encoding/json"
	"lab1/models"
	"log"
	"net"
	"sort"
)

const (
	baseHost = "185.102.139.168"
	basePort = "50051"
)

func handleClient(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req models.SortRequest
		err := decoder.Decode(&req)
		if err != nil {
			log.Println("Error in connecting:", err)
			sendError(encoder, "Invalid request format")
			return
		}

		sortedArray := make([]int, len(req.Array))
		copy(sortedArray, req.Array)
		if req.Order == "asc" {
			sort.Ints(sortedArray)
		} else {
			sort.Sort(sort.Reverse(sort.IntSlice(sortedArray)))
		}

		res := models.SortResponse{
			OriginalArray: req.Array,
			SortedArray:   sortedArray,
		}

		err = encoder.Encode(&res)
		if err != nil {
			log.Println("Error in sending request:", err)
		}

		log.Printf("Array is sorted: %v -> %v", req.Array, sortedArray)
	}
}

func sendError(encoder *json.Encoder, message string) {
	errMsg := models.ErrorMessage{
		Message: message,
	}
	encoder.Encode(&errMsg)
}

func main() {
	address := net.JoinHostPort(baseHost, basePort)

	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Can't start server:", err)
	}
	defer ln.Close()
	log.Printf("Server listenning on addres: %s\n", address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Connecting error:", err)
			continue
		}

		log.Println("New Connection")
		go handleClient(conn)
	}
}
