package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/testcontainers/helloworld/internal/util"
)

// StartServing starts serving content on ports 8080 and 8081
func StartServing() {

	// We will delay, if configured to do so with an environment variable
	delayStart := util.GetEnvInt("DELAY_START_MSEC", 0)
	log.Printf("DELAY_START_MSEC: %d\n", delayStart)

	// Create a UUID for this container instance
	instanceUUID := uuid.New().String()

	// start both servers, with delay before each
	startServerOnPort(8080, instanceUUID, delayStart)
	startServerOnPort(8081, instanceUUID, delayStart)
}

func startServerOnPort(port int, instanceUUID string, delayStart int) {
	fs := http.FileServer(http.Dir("./static"))

	server := http.NewServeMux()
	server.Handle("/", fs)
	server.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "PONG")
	})
	server.HandleFunc("/uuid", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, instanceUUID)
	})
	server.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {

		for name, value := range r.Header {
			fmt.Fprintf(w, "%s:%s\n", name, value)
		}
		fmt.Fprintf(w, "\n\n")

		host := r.Header.Get("X-Forwarded-Host")
		if host == "" {
			host, _, _ = net.SplitHostPort(r.Host)
		}
		scheme := r.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			if r.TLS == nil {
				scheme = "http"
			} else {
				scheme = "https"
			}
		}
		port := r.Header.Get("X-Forwarded-Port")
		if port == "" {
			_, port, _ = net.SplitHostPort(r.Host)
		}
		path := r.URL.Path
		fmt.Fprintf(w, "%s://%s:%s%s", scheme, host, port, path)
	})

	// Delay before the server starts
	log.Printf("Sleeping for %d ms", delayStart)
	time.Sleep(time.Duration(delayStart) * time.Millisecond)

	log.Printf("Starting server on port %d", port)
	portName := fmt.Sprintf(":%d", port)
	go func() {
		log.Fatal(http.ListenAndServe(portName, handlers.LoggingHandler(os.Stdout, server)))
	}()
}
