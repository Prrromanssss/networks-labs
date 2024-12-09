package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jlaffaye/ftp"
)

// Configuring the WebSocket upgrader with an origin check
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Structure to manage client data, including WebSocket and FTP connection
type client struct {
	conn      *websocket.Conn
	ftpConn   *ftp.ServerConn
	connected bool
}

// Map to manage all active clients
var clients = make(map[*websocket.Conn]*client)

func main() {
	// HTTP handler for WebSocket connections
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Starting WebSocket server on port 55187...")
	log.Fatal(http.ListenAndServe(":55187", nil))
}

// Handles WebSocket connections and processes FTP commands
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error establishing WebSocket connection:", err)
		return
	}
	defer wsConn.Close()

	c, exists := clients[wsConn]
	if !exists {
		c = &client{conn: wsConn, connected: false}
		clients[wsConn] = c
	}

	var ftpConn *ftp.ServerConn

	for {
		// Read messages from the WebSocket connection
		_, msg, err := wsConn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		command := string(msg)
		args := strings.Fields(command)

		if len(args) == 0 {
			wsConn.WriteMessage(websocket.TextMessage, []byte("Invalid command."))
			continue
		}

		fmt.Println("Connect", wsConn.RemoteAddr(), args)

		// Handle FTP connection
		if len(args) == 4 && args[0] == "connect" {
			ftpHost := args[1]
			ftpLogin := args[2]
			ftpPassword := args[3]

			// Attempt to establish an FTP connection
			ftpConn, err = ftp.Dial(ftpHost, ftp.DialWithTimeout(5*time.Second))
			if err != nil {
				wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to connect to server: %v", err)))
				continue
			}

			// Attempt to log in to the FTP server
			err = ftpConn.Login(ftpLogin, ftpPassword)
			if err != nil {
				wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Login failed: %v", err)))
				continue
			}

			c.ftpConn = ftpConn
			c.connected = true
			wsConn.WriteMessage(websocket.TextMessage, []byte("Successfully connected to the FTP server!"))
			continue
		}

		// Ensure the client is connected before processing further commands
		if !c.connected {
			wsConn.WriteMessage(websocket.TextMessage, []byte("Please connect to the FTP server first."))
			continue
		}

		// Handle various FTP commands
		switch args[0] {
		case "cd":
			if len(args) < 2 {
				wsConn.WriteMessage(websocket.TextMessage, []byte("Specify the path to change the directory."))
				continue
			}
			changeDir(ftpConn, args[1], wsConn)
		case "mkdir":
			if len(args) < 2 {
				fmt.Println("Specify the path to create a directory.")
				continue
			}
			createDir(ftpConn, args[1], wsConn)
		case "ls":
			if len(args) < 2 {
				wsConn.WriteMessage(websocket.TextMessage, []byte("Specify the path to list the directory contents."))
				continue
			}
			listFiles(ftpConn, args[1], wsConn)
		case "rmEmptyDir":
			if len(args) < 2 {
				wsConn.WriteMessage(websocket.TextMessage, []byte("Specify the path to the empty directory to delete."))
				continue
			}
			removeEmptyDir(c.ftpConn, args[1], wsConn)
		case "rmRecursiveDir":
			if len(args) < 2 {
				wsConn.WriteMessage(websocket.TextMessage, []byte("Specify the path to the directory for recursive deletion."))
				continue
			}
			removeDirRecursive(c.ftpConn, args[1], wsConn)
		case "rmFile":
			if len(args) < 2 {
				wsConn.WriteMessage(websocket.TextMessage, []byte("Specify the path to the file to delete."))
				continue
			}
			removeFile(c.ftpConn, args[1], wsConn)
		case "exit":
			wsConn.WriteMessage(websocket.TextMessage, []byte("Exiting program..."))
			return
		default:
			wsConn.WriteMessage(websocket.TextMessage, []byte("Unknown command."))
		}
	}
}

// Create a new directory on the FTP server
func createDir(conn *ftp.ServerConn, dir string, wsConn *websocket.Conn) {
	err := conn.MakeDir(dir)
	if err != nil {
		log.Printf("Failed to create directory %s: %v", dir, err)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to create directory %s: %v", dir, err)))
	} else {
		log.Printf("Directory %s created successfully.", dir)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Directory %s created successfully.", dir)))
	}
}

// Change the working directory on the FTP server
func changeDir(conn *ftp.ServerConn, dir string, wsConn *websocket.Conn) {
	err := conn.ChangeDir(dir)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to change to directory %s: %v", dir, err)))
	} else {
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Successfully changed to directory: %s", dir)))
	}
}

// List the contents of a directory on the FTP server
func listFiles(conn *ftp.ServerConn, dir string, wsConn *websocket.Conn) {
	entries, err := conn.List(dir)
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to list directory contents %s: %v", dir, err)))
		return
	}

	var files []string
	for _, entry := range entries {
		var color string
		if entry.Type == ftp.EntryTypeFolder {
			color = "blue"
		} else {
			color = "green"
		}

		files = append(files, fmt.Sprintf(`<span style="color:%s">%s</span>`, color, entry.Name))
	}
	wsConn.WriteMessage(websocket.TextMessage, []byte("<br>"+strings.Join(files, "<br>")))
}

// Remove an empty directory on the FTP server
func removeEmptyDir(conn *ftp.ServerConn, dir string, wsConn *websocket.Conn) {
	err := conn.RemoveDir(dir)
	if err != nil {
		log.Printf("Failed to remove empty directory %s: %v", dir, err)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to remove empty directory %s: %v", dir, err)))
	} else {
		log.Printf("Empty directory %s removed successfully.", dir)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Empty directory %s removed successfully.", dir)))
	}
}

// Recursively remove a directory and its contents on the FTP server
func removeDirRecursive(conn *ftp.ServerConn, dir string, wsConn *websocket.Conn) {
	entries, err := conn.List(dir)
	if err != nil {
		log.Printf("Failed to list directory contents %s: %v", dir, err)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to list directory contents %s: %v", dir, err)))
		return
	}

	for _, entry := range entries {
		fullPath := dir + "/" + entry.Name
		if entry.Name == "." || entry.Name == ".." {
			log.Printf("Skipping directory %s", fullPath)
			continue
		}

		if entry.Type == ftp.EntryTypeFile {
			err := conn.Delete(fullPath)
			if err != nil {
				log.Printf("Failed to delete file %s: %v", fullPath, err)
				wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to delete file %s: %v", fullPath, err)))
			} else {
				log.Printf("File %s deleted successfully.", fullPath)
				wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("File %s deleted successfully.", fullPath)))
			}
		} else if entry.Type == ftp.EntryTypeFolder {
			log.Printf("Recursively processing subdirectory: %s", fullPath)
			removeDirRecursive(conn, fullPath, wsConn)
		}
	}

	err = conn.RemoveDir(dir)
	if err != nil {
		log.Printf("Failed to remove directory %s: %v", dir, err)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to remove directory %s: %v", dir, err)))
	} else {
		log.Printf("Directory %s removed successfully.", dir)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Directory %s removed successfully.", dir)))
	}
}

// Remove a file on the FTP server
func removeFile(conn *ftp.ServerConn, file string, wsConn *websocket.Conn) {
	err := conn.Delete(file)
	if err != nil {
		log.Printf("Failed to delete file %s: %v", file, err)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Failed to delete file %s: %v", file, err)))
	} else {
		log.Printf("File %s deleted successfully.", file)
		wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("File %s deleted successfully.", file)))
	}
}
