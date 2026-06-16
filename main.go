package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."

	// port "80" is the default http port, but ports 1-1023 are protected so 8080 is a standard alternative
	const port = "8080"

	// "mux" is short for "multiplexer"
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	// "server" is a pointer to an "http.Server" struct
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
