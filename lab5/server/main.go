package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mmcdole/gofeed"
)

const (
	rssUrl = "https://lenta.ru/rss"
	dsn    = "iu9networkslabs:Je2dTYr6@tcp(students.yss.su:3306)/iu9networkslabs"
)

type Item struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

func transliterate(text string) string {
	translitMap := map[rune]string{
		'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo", 'Ж': "Zh", 'З': "Z",
		'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N", 'О': "O", 'П': "P", 'Р': "R",
		'С': "S", 'Т': "T", 'У': "U", 'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch", 'Ш': "Sh", 'Щ': "Sch",
		'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu", 'Я': "Ya",
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z",
		'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r",
		'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	}

	var sb strings.Builder
	for _, r := range text {
		if translit, ok := translitMap[r]; ok {
			sb.WriteString(translit)
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func parseRSS(db *sql.DB) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssUrl)
	if err != nil {
		log.Fatalf("Error parsing RSS: %v", err)
	}

	for _, item := range feed.Items {
		transliteratedTitle := transliterate(item.Title)

		description := item.Description
		if description == "" {
			description = getFirstThreeWords(transliteratedTitle)
		}

		fmt.Println("Title:", transliteratedTitle)
		fmt.Println("Text:", description)

		_, err := db.Exec(`
			INSERT INTO iu9boykoroman (title, text) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE text = ?`, transliteratedTitle, description, description)
		if err != nil {
			log.Printf("Error inserting/updating article: %v", err)
		}
	}

	log.Println("Successfully parsed and updated iu9boykoroman in database")
}

func getFirstThreeWords(title string) string {
	words := strings.Fields(title)
	if len(words) > 3 {
		return strings.Join(words[:3], " ")
	}
	return title
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, text FROM iu9boykoroman")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	articles := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Title, &item.Text); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		articles = append(articles, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articles)
}

func main() {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	parseRSS(db)

	http.HandleFunc("/articles", articlesHandler)

	go func() {
		log.Println("Starting RSS server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gracefully...")
}
