package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type OilPrice struct {
	Name          string
	Value         string
	Change        string
	ChangePercent string
}

func parseOilPricesFromHTML(htmlContent string) ([]OilPrice, error) {
	var prices []OilPrice

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге HTML: %w", err)
	}

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var price OilPrice
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "td" {
					class := getClass(c)
					switch {
					case strings.Contains(class, "blend_name"):
						price.Name = extractText(c)
					case strings.Contains(class, "value"):
						price.Value = extractText(c)
					case strings.Contains(class, "change_amount"):
						price.Change = extractText(c)
					case strings.Contains(class, "change_percent"):
						price.ChangePercent = extractText(c)
					}
				}
			}
			if price.Name != "" && price.Value != "" {
				prices = append(prices, price)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	return prices, nil
}

func getClass(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			return attr.Val
		}
	}
	return ""
}

func extractText(n *html.Node) string {
	var sb strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(sb.String())
}

func handler(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("https://oilprice.com/")
	if err != nil {
		log.Printf("Ошибка при получении данных: %v", err)
		http.Error(w, "Не удалось получить данные", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	htmlContent, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Ошибка при чтении данных: %v", err)
		http.Error(w, "Не удалось прочитать данные", http.StatusInternalServerError)
		return
	}

	prices, err := parseOilPricesFromHTML(string(htmlContent))
	if err != nil {
		log.Printf("Ошибка при парсинге данных: %v", err)
		http.Error(w, "Не удалось распарсить данные", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "<html><head><title>Цены на нефть</title></head><body>")
	fmt.Fprintf(w, "<h1>Текущие цены на нефть</h1>")
	fmt.Fprintf(w, "<ul>")
	for _, price := range prices {
		fmt.Fprintf(w, "<li><strong>%s:</strong> %s (изменение: %s, %% изменения: %s)</li>",
			price.Name, price.Value, price.Change, price.ChangePercent)
	}
	fmt.Fprintf(w, "</ul></body></html>")
}

func main() {
	log.Println("Запуск сервера на порту 8080")
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
