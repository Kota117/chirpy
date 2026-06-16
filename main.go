package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot string = "."

	// port "80" is the default http port, but ports 1-1023 are protected so 8080 is a standard alternative
	const port string = "8080"

	// "mux" is short for "multiplexer"
	var mux *http.ServeMux = http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	// "server" is a pointer to an "http.Server" struct
	var server *http.Server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
