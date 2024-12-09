package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	db *sql.DB
)

type Item struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected")

	for {
		articles, err := fetchArticles()
		if err != nil {
			log.Println("Error fetching articles:", err)
			return
		}

		if err := conn.WriteJSON(articles); err != nil {
			log.Println("Error sending articles:", err)
			return
		}

		time.Sleep(5 * time.Second)
	}
}

func fetchArticles() ([]Item, error) {
	rows, err := db.Query("SELECT id, title, text FROM iu9boykoroman")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Title, &item.Text); err != nil {
			return nil, err
		}
		articles = append(articles, item)
	}
	return articles, nil
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := fetchArticles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articles)
}

func main() {
	var err error
	db, err = sql.Open("mysql", "iu9networkslabs:Je2dTYr6@tcp(students.yss.su)/iu9networkslabs")
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/articles", articlesHandler)

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
