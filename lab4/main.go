package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling tunneling for host: %s", r.Host)
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("Error connecting to destination: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	log.Printf("Connected to destination: %s", r.Host)
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Error hijacking connection: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	log.Println("Hijacked client connection")

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	log.Println("Starting data transfer")
	defer destination.Close()
	defer source.Close()
	bytesCopied, err := io.Copy(destination, source)
	if err != nil {
		log.Printf("Error during transfer: %v", err)
	}
	log.Printf("Transfer complete, bytes copied: %d", bytesCopied)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("Handling HTTP request: %s %s", req.Method, req.URL)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Printf("Error processing request: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	bytesCopied, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
	log.Printf("HTTP request handled, bytes copied: %d", bytesCopied)
}

func copyHeader(dst, src http.Header) {
	log.Println("Copying headers")
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func main() {
	var pemPath string
	flag.StringVar(&pemPath, "pem", "server.pem", "path to pem file")
	var keyPath string
	flag.StringVar(&keyPath, "key", "server.key", "path to key file")
	var proto string
	flag.StringVar(&proto, "proto", "https", "Proxy protocol (http or https)")
	flag.Parse()

	log.Printf("Starting proxy server with protocol: %s", proto)

	if proto != "http" && proto != "https" {
		log.Fatal("Protocol must be either http or https")
	}

	server := &http.Server{
		Addr: ":50051",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Received request: %s %s", r.Method, r.URL)
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	if proto == "http" {
		log.Println("Starting HTTP server on :50051")
		log.Fatal(server.ListenAndServe())
	} else {
		log.Println("Starting HTTPS server on :50051")
		log.Fatal(server.ListenAndServeTLS(pemPath, keyPath))
	}
}
