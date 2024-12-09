package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	host = "http://185.102.139.169"
	port = "8081"
)

func modifyLinksWithGoQuery(htmlContent string, baseURL *url.URL) (string, error) {
	var proxyBaseURL = net.JoinHostPort(host, port)

	// HTML -> GoQuery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	// Fix links.
	toAbsoluteURL := func(link string) string {
		u, err := url.Parse(link)
		if err != nil {
			return link
		}
		return baseURL.ResolveReference(u).String()
	}

	// Replace all links in tag <a href>.
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			absoluteHref := toAbsoluteURL(href)
			newHref := fmt.Sprintf("%s/?url=%s", proxyBaseURL, absoluteHref)
			s.SetAttr("href", newHref)
		}
	})

	// Replace all links in tag <img src>.
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			absoluteSrc := toAbsoluteURL(src)
			newSrc := fmt.Sprintf("%s/?url=%s", proxyBaseURL, absoluteSrc)
			s.SetAttr("src", newSrc)
		}
	})

	// Replace all links in tag <link href>.
	doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			absoluteHref := toAbsoluteURL(href)
			newHref := fmt.Sprintf("%s/?url=%s", proxyBaseURL, absoluteHref)
			s.SetAttr("href", newHref)
		}
	})

	// Replace all links in tag <script src>.
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			absoluteSrc := toAbsoluteURL(src)
			newSrc := fmt.Sprintf("%s/?url=%s", proxyBaseURL, absoluteSrc)
			s.SetAttr("src", newSrc)
		}
	})

	// GoQuery -> HTML
	html, err := doc.Html()
	if err != nil {
		return "", err
	}

	return html, nil
}

func handleProxy(w http.ResponseWriter, r *http.Request) {

	// Get url.
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	// Check url.
	parsedURL, err := url.Parse(targetURL)
	if err != nil || !parsedURL.IsAbs() {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Request for url.
	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Error making request to the target URL", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response from the target URL", http.StatusInternalServerError)
		return
	}

	// Handle only html, replace all links.
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		modifiedBody, err := modifyLinksWithGoQuery(string(body), parsedURL)
		if err != nil {
			http.Error(w, "Error processing HTML", http.StatusInternalServerError)
			return
		}
		body = []byte(modifiedBody)
	}

	// Copy response for the client.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	http.HandleFunc("/", handleProxy)

	log.Println("Proxy server starts on port: %s", port)

	log.Panic(http.ListenAndServe(
		fmt.Sprintf(":%s", port),
		nil,
	))
}
