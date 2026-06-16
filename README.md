# Chirpy

A fully-featured social media platform API written in Go.

## Prerequisites

- **Go**: Version 1.22+ installed on the local machine.
- **curl**: For manually testing API endpoints via the terminal.

## Setup

Before running the server for the first time, initialize the Go module:

```bash
go mod init github.com/Kota117/chirpy
```

## Features
- **Static File Serving**: Serves HTML, CSS, and JavaScript assets from the root directory using Go's standard library `http.FileServer`.

## Project Structure
```text
.
├── main.go      # Entry point for the Go server
├── index.html   # Root HTML file served at http://localhost:8080
└── go.mod       # Go module definition
```

## Running the Server

To build and run the server locally on port `8080`, run the following command in a terminal from the root directory of the project:

```bash
go build -o out && ./out
```

* `go build -o out`: Compiles the Go source code into an executable binary named `out`.
* `&&`: A bash logical AND operator. Ensures `./out` runs if and only if the compilation step succeeds (exits with status code 0).
* `./out`: Executes the newly compiled binary.

*Note: Go is a compiled language, so the server will not automaticcally reflect code changes. The server must be stopped with `Ctrl+C`, rebuilt with the above command, and restarted whenever changes are made.*

## Testing the Server
While the server is running in one terminal window, it can be manually tested in another terminal window or in a browser at `http://localhost:8080`.

### GET Request
Use `curl` to send a `GET` request and inspect the response headers:

```bash
curl -i http://localhost:8080/
```

* `-i`/`--include`: Tells `curl` to print the HTTP response headers (like `HTTP/1.1 404 Not Found`) along with any body content.
