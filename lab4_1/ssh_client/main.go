package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = username[:len(username)-1]

	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = password[:len(password)-1]

	fmt.Print("Enter host: ")
	host, _ := reader.ReadString('\n')
	host = host[:len(host)-1]

	fmt.Print("Enter port: ")
	port, _ := reader.ReadString('\n')
	port = port[:len(port)-1]

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	fmt.Println("Connected to SSH server. Type 'exit' to disconnect.")

	for {
		fmt.Print("Enter command: ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		if command == "exit" {
			fmt.Println("Disconnecting...")
			break
		}

		session, err := client.NewSession()
		if err != nil {
			log.Println("Failed to create session: ", err)
			continue
		}

		output, err := session.CombinedOutput(command)
		if err != nil {
			log.Println("Failed to run command: ", err)
		}

		fmt.Println(string(output))

		session.Close()
	}
}
