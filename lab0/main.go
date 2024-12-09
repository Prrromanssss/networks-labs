package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
)

const (
	baseHost = "localhost"
	basePort = "9035"
)

const (
	rssUrl = "https://lenta.ru/rss"
)

func RssHandler(w http.ResponseWriter, r *http.Request) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(rssUrl)
	if err != nil {
		fmt.Fprintf(w, "Error whie getting the RSS: %v", err)
		return
	}

	fmt.Fprintf(w, "<h1>%s</h1>", feed.Title)
	fmt.Fprintf(w, "<p>%s</p>", feed.Description)

	for _, item := range feed.Items {
		fmt.Fprintf(w, "<h3><a href='%s'>%s</a></h3>", item.Link, item.Title)
		fmt.Fprintf(w, "<p>%s</p>", item.Description)
	}
}

func BaseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<h1>Hello, you can see rss of lenta on <a href="/rss">rss<a><h1>`)
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodColor := color.New(color.FgGreen).SprintFunc()
		pathColor := color.New(color.FgCyan).SprintFunc()
		log.Printf("Method: %s, Path: %s", methodColor(r.Method), pathColor(r.URL.Path))

		next.ServeHTTP(w, r)

	})
}

func main() {
	serverAddress := net.JoinHostPort(baseHost, basePort)
	server := &http.Server{
		Addr:    serverAddress,
		Handler: nil,
	}

	http.Handle("/rss", loggerMiddleware(http.HandlerFunc(RssHandler)))
	http.Handle("/", loggerMiddleware(http.HandlerFunc(BaseHandler)))

	go func() {
		infoColor := color.New(color.FgYellow).SprintFunc()
		log.Printf("%s Server listening on address: %s", infoColor("INFO:"), fmt.Sprintf("http://%s", serverAddress))

		if err := server.ListenAndServe(); err != nil {
			errorColor := color.New(color.FgRed).SprintFunc()
			log.Printf("%s Server Error: %v\n", errorColor("ERROR:"), err)
		}
	}()

	gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	shutdownColor := color.New(color.FgMagenta).SprintFunc()
	log.Println(shutdownColor("Shutting down server..."))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		errorColor := color.New(color.FgRed).SprintFunc()
		log.Panicf("%s Server forced to shutdown: %v", errorColor("ERROR:"), err)
	}

	successColor := color.New(color.FgGreen).SprintFunc()
	log.Println(successColor("Server exiting gracefully"))
}
