package main

import (
	"log"
	"net/http"
)

func main() {
	// port "80" is the default http port, but ports 1-1023 are protected so 8080 is a standard alternative
	const port = "8080"

	// "mux" is short for "multiplexer"
	mux := http.NewServeMux()

	// "server" is a pointer to an "http.Server" struct
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
