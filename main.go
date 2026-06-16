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

	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.HandleFunc("/healthz", handlerReadiness)

	var server *http.Server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
