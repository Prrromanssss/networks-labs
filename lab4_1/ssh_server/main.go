package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gliderlabs/ssh"
)

const (
	validUser     = "user"
	validPassword = "password"
)

func main() {
	ssh.Handle(func(s ssh.Session) {
		command := s.Command()
		if len(command) == 0 {
			s.Write([]byte("No command provided\n"))
			return
		}

		log.Printf("Executing command: %s", command)

		switch command[0] {
		case "mkdir":
			if len(command) > 1 {
				err := os.Mkdir(command[1], 0755)
				if err != nil {
					s.Write([]byte(fmt.Sprintf("Error creating directory: %s\n", err)))
				} else {
					s.Write([]byte(fmt.Sprintf("Directory %s created\n", command[1])))
				}
			} else {
				s.Write([]byte("Usage: mkdir <directory>\n"))
			}
		case "rmdir":
			if len(command) > 1 {
				err := os.Remove(command[1])
				if err != nil {
					s.Write([]byte(fmt.Sprintf("Error removing directory: %s\n", err)))
				} else {
					s.Write([]byte(fmt.Sprintf("Directory %s removed\n", command[1])))
				}
			} else {
				s.Write([]byte("Usage: rmdir <directory>\n"))
			}
		case "ls":
			if len(command) > 1 {
				files, err := os.ReadDir(command[1])
				if err != nil {
					s.Write([]byte(fmt.Sprintf("Error listing directory: %s\n", err)))
				} else {
					for _, file := range files {
						s.Write([]byte(file.Name() + "\n"))
					}
				}
			} else {
				s.Write([]byte("Usage: ls <directory>\n"))
			}
		case "mv":
			if len(command) > 2 {
				err := os.Rename(command[1], command[2])
				if err != nil {
					s.Write([]byte(fmt.Sprintf("Error moving file: %s\n", err)))
				} else {
					s.Write([]byte(fmt.Sprintf("Moved %s to %s\n", command[1], command[2])))
				}
			} else {
				s.Write([]byte("Usage: mv <source> <destination>\n"))
			}
		case "rm":
			if len(command) > 1 {
				err := os.Remove(command[1])
				if err != nil {
					s.Write([]byte(fmt.Sprintf("Error removing file: %s\n", err)))
				} else {
					s.Write([]byte(fmt.Sprintf("File %s removed\n", command[1])))
				}
			} else {
				s.Write([]byte("Usage: rm <filename>\n"))
			}
		case "ping":
			if len(command) > 1 {
				out, err := exec.Command("ping", "-c", "4", command[1]).Output()
				if err != nil {
					s.Write([]byte(fmt.Sprintf("Error executing ping: %s\n", err)))
				} else {
					s.Write(out)
				}
			} else {
				s.Write([]byte("Usage: ping <hostname>\n"))
			}
		default:
			s.Write([]byte("Unknown command\n"))
		}
	})

	log.Println("Starting SSH server on port 2222...")
	if err := ssh.ListenAndServe(":2222", nil,
		ssh.PasswordAuth(func(ctx ssh.Context, password string) bool {
			log.Printf("Login attempt: User: %s", ctx.User())
			if ctx.User() == validUser && password == validPassword {
				log.Printf("Authentication successful for user: %s", ctx.User())
				return true
			}
			log.Printf("Failed login attempt for user: %s", ctx.User())
			return false
		}),
	); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
