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
- **Static File Serving**: Serves HTML and media assets from the `/app/` path using `http.FileServer` and `http.StripPrefix`.
- **Health Check Endpoint**: Includes a lightweight readiness endpoint at `GET /healthz` to verify server availability.
- **Request Metrics**: Tracks the number of file server hits using an `atomic.Int32` counter. Accessible at `GET /metrics`.  
*Note: The request counter is stored in memory and resets to 0 whenever the server is stopped and restarted.*
- **Metrics Reset**: Resets the hit counter back to zero via the `POST /reset` endpoint.


## Project Structure
```text
.
├── assets/      # Static assets like images and logos
│   └── logo.png
├── .gitignore   # Disables version-tracking for any included files/folders
├── main.go      # Entry point for the Go server
├── metrics.go   # Handler for getting the number of requests since the server was last started
├── readiness.go # Handler for testing if the server is up and ready to receive traffic
├── reset.go     # Handler for resetting the request counter
├── index.html   # Root HTML file served at http://localhost:8080
└── go.mod       # Go module definition
```


## Usage
To serve additional media assets, place them in the assets/ directory. They will be automatically available at `http://localhost:8080/app/assets/<filename>`.


## Running the Server
To build and run the server locally on port `8080`, run the following command in a terminal from the root directory of the project:

```bash
go build -o out && ./out
```

* `go build -o out`: Compiles the Go source code into an executable binary named `out`.
* `&&`: A bash logical AND operator. Ensures `./out` runs if and only if the compilation step succeeds (exits with status code 0).
* `./out`: Executes the newly compiled binary.

*Note: Per standard practice, the compiled `out` binary is not version-tracked.*  
*Note: Go is a compiled language, so the server will not automatically reflect code changes. The server must be stopped with `Ctrl+C`, rebuilt with the above command, and restarted whenever changes are made.*


## Testing the Server
While the server is running in one terminal window, it can be manually tested in another terminal window. Alternatively, the entire breadth of content can be viewed in a browser at `http://localhost:8080/app/`.

### Inspect index.html
```bash
curl -i http://localhost:8080/app/
```

* `-i`/`--include`: Tells `curl` to print the HTTP response headers (like `HTTP/1.1 404 Not Found`) along with any body content.

### Inspect Media
```bash
curl -I http://localhost:8080/app/assets/logo.png
```

* `-I`/`--head`: Tells `curl` to print only the HTTP response headers without any body content

### Inspect the Health Endpoint
To check if the server is up and ready to receive traffic (only accepts `GET` requests):

```bash
curl -i http://localhost:8080/healthz
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 2

OK
```

The server will reject any HTTP method other than `GET` at this endpoint:

```bash
curl -i -X POST http://localhost:8080/healthz
```

Expected response:
```text
HTTP/1.1 405 Method Not Allowed
...
```

### Check how many requests have been served
```bash
curl -i http://localhost:8080/metrics
```

Expected response:
```text
Hits: 3
```
*Note: the `3` is expected to be any positive integer representing the count of requests served since the server was last started*

### Reset the request counter
Only accepts `POST` requests:

```bash
curl -i -X POST http://localhost:8080/reset
```

Expected response:
```text
Hits reset to 0
```

*Note: Sending any non-`POST` HTTP request to `/reset` will result in a `405 Method Not Allowed` response.*
