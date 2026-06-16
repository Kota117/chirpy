# Chirpy

A fully-featured social media platform API written in Go.

## Prerequisites
- **Go**: Version 1.22+ installed on the local machine.
- **curl**: For manually testing API endpoints via the terminal.

## Setup
After cloning the repository (`https://github.com/Kota117/chirpy`), it is recommended to run the following to ensure the `go.mod` file matches the source code:

```bash
go mod tidy
```

## Features
- **Static File Serving**: Serves HTML and media assets from the root directory using Go's standard library (`http.FileServer`).

## Project Structure
```text
.
├── assets/      # Static assets like images and logos
│   └── logo.png
├── main.go      # Entry point for the Go server
├── index.html   # Root HTML file served at http://localhost:8080
└── go.mod       # Go module definition
```

## Usage
To serve additional media assets, place them in the assets/ directory. They will be automatically available at `http://localhost:8080/assets/<filename>`.

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
While the server is running in one terminal window, it can be manually tested in another terminal window. Alternatively, the entire breadth of content can be viewed in a browser at `http://localhost:8080`.

### Inspect index.html
```bash
curl -i http://localhost:8080/
```

* `-i`/`--include`: Tells `curl` to print the HTTP response headers (like `HTTP/1.1 404 Not Found`) along with any body content.

### Inspect Media
```bash
curl -I http://localhost:8080/assets/logo.png
```

* `-I`/`--head`: Tells `curl` to print only the HTTP response headers without any body content