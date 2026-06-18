# Chirpy

A fully-featured social media platform API written in Go.


## Prerequisites
- **Go**: Version 1.22+ installed on the local machine.
- **curl**: For manually testing API endpoints via the terminal.
- **PostgreSQL**: Version 15+ (installed via WSL/Ubuntu).
- **Goose**: For running database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`).
- **SQLC**: For generating type-safe Go code from SQL queries (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`).


## Architecture
Chirpy follows a monolithic structure but maintains a clean separation between the user-facing application, the data API, and administrative tooling by using `/app`, `/api`, and `/admin` namespaces.  
*Note: The `/admin` namespace is **not** inherently more secure than the others, it is simply an organizational structure.*


## Features
- **Static File Serving**: Serves HTML and media assets from the `/app/` path using `http.FileServer` and `http.StripPrefix`.  
*Note: The `http.StripPrefix` allows the file system to remain agnostic of the URL structure.*
- **Health Check Endpoint**: Includes a lightweight readiness endpoint at `GET /api/healthz` to verify server availability.
- **Request Metrics**: Tracks the number of file server hits using an `atomic.Int32` counter. Accessible via the `GET /admin/metrics` endpoint.  
*Note: The request counter is stored in memory and resets to 0 whenever the server is stopped and restarted.*
- **Metrics Reset**: Resets the hit counter back to zero and deletes all users from the database via `POST /admin/reset`. To gate this dangerous endpoint, it is only accessible when `PLATFORM=dev`; returns `403 Forbidden` otherwise.
- **User Creation**: Creates a new user via `POST /api/users`. Accepts an `email` in the JSON request body and returns the user's `id`, `created_at`, `updated_at`, and `email`.


## Project Structure
```text
.
├── assets/                  # Static assets like images and logos
│   └── logo.png
├── internal/
│   └── database/            # SQLC-generated Go database code
├── sql/                     
│   ├── queries/             # SQLC query definitions
│   │   └── users.sql
│   └── schema/              # Goose migration files
│       └── 001_users.sql
├── .env                     # Local environment variables (not version-tracked)
├── .gitignore               # Disables version-tracking for any included files/folders
├── go.mod                   # Go module definition
├── handler_metrics.go       # Handler for getting the number of requests since the server was last started
├── handler_readiness.go     # Handler for testing if the server is up and ready to receive traffic
├── handler_reset.go         # Handler for resetting the request counter
├── handler_validate.go      # Handler for validating Chirp content
├── handler_users_create.go  # Handler for creating a new user
├── index.html               # Root HTML file served at http://localhost:8080
├── json.go                  # Shared helpers for encoding JSON responses and errors
├── main.go                  # Entry point for the Go server
└── sqlc.yaml                # SQLC configuration
```


## Setup
After cloning the repository (`https://github.com/Kota117/chirpy`), it is recommended to run the following to ensure the `go.mod` file matches the source code:

```bash
go mod tidy
```


## Dependencies
Each dependency can be added via `go get <name>`:
- `github.com/google/uuid` — UUID generation for database records
- `github.com/lib/pq` — PostgreSQL driver for `database/sql`
- `github.com/joho/godotenv` — Loads environment variables from a `.env` file


## Database Setup (WSL / Ubuntu)
Chirpy uses a PostgreSQL database for persistent data storage. The database can be set up and configured inside a WSL environment via the following steps.

### 1. Installation
Update the system's package list and install PostgreSQL along with its contrib utilities:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

Verify the installation and check the version:
```bash
psql --version
```

### 2. Set System Password
Set a password for the WSL system user `postgres` (the password can be set to `postgres` for simplicity):
```bash
sudo passwd postgres
```

### 3. Start the PostgreSQL Service
PostgreSQL does not start automatically on WSL. It must be started manually when development begins:
```bash
sudo service postgresql start
```
*Note: To stop the service later, run `sudo service postgresql stop`*

### 4. Create the Database and User
Access the PostgreSQL interactive terminal (`psql`) as the superuser `postgres`:
```bash
sudo -u postgres psql
```
Once inside the `postgres=#` prompt, run the following SQL queries:

1. Create the application database:
    ```sql
    CREATE DATABASE chirpy;
    ```

2. Set the database password for the `postgres` user:
    ```sql
    ALTER USER postgres WITH PASSWORD 'postgres';
    ```

3. Verify the connection to the new database:
    ```bash
    \c chirpy
    ```
    *Note: The prompt should change to chirpy=#*

4. Exit the psql shell:
    ```bash
    \q
    ```


## Environment Configuration
Chirpy reads configuration from a `.env` file in the project root. This file is **not** version-tracked.

Create a `.env` file with the following:
```bash
DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
PLATFORM="<dev|prod>"
```

Example (Linux/WSL):
```bash
DB_URL="postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
```


## Database Migrations
Chirpy uses [Goose](https://github.com/pressly/goose) to manage database schema migrations. Migration files live in `sql/schema/` and are plain `.sql` files with special Goose comments.

### Running Migrations
To upgrade the database to new schema, `cd` into the `sql/schema` directory, then run:
```bash
goose postgres "<connection_string>" up
```

To revert the most recent migration:
```bash
goose postgres "<connection_string>" down
```

**Connection string format:**
```text
postgres://username:password@host:port/database
```

Example (Linux/WSL):
```bash
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" up
```

**Verify the migration applied successfully:**
```bash
psql "<connection_string>"
\dt
```

Example (Linux/WSL):
```bash
psql "postgres://postgres:postgres@localhost:5432/chirpy"
\dt
```


## Current Schema
| Table |	Column      | Type      |	Constraints      |
| ----- | ----------- | --------- | ---------------- |
| users	| id	        | UUID    	| PRIMARY KEY      |
| users	| created_at	| TIMESTAMP	| NOT NULL         |
| users	| updated_at	| TIMESTAMP |	NOT NULL         |
| users	| email	      | TEXT    	| NOT NULL, UNIQUE |


## Generating Database Code (SQLC)
Chirpy uses [SQLC](https://sqlc.dev/) to generate type-safe Go code from SQL queries.

To regenerate the `internal/database` package after modifying queries in `sql/queries/`:
```bash
sqlc generate
```
*Note: This command should be run from the `root` of the project.*

## Usage
| Endpoint              | Method | Description                                                   |
| --------------------- | ------ | ------------------------------------------------------------- |
| `/app/*`              | GET    | Serves static frontend files                                  |
| `/api/healthz`        | GET    | Readiness check                                               |
| `/api/users`          | POST   | Create new user                                               |
| `/api/validate_chirp` | POST   | Validate a Chirp (max 140 chars, profanity filtered)          |
| `/admin/metrics`      | GET    | Retrieve hit counter (HTML)                                   |
| `/admin/reset`        | POST   | Reset hit counter and delete all users (dev environment only) |


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


## Manually Testing the Server
While the server is running in one terminal window, the back-end can be manually tested in another terminal window. Additionally, the front-end content can be viewed in a browser at `http://localhost:8080/app/`.  
  
Admin metrics can be viewed at `http://localhost:8080/admin/metrics`.

### Inspect the contents of index.html
```bash
curl -i http://localhost:8080/app/
```
* `-i`/`--include`: Tells `curl` to print the HTTP response headers along with any body content.
*Note: The default HTTP method used by curl is `GET`.*

### Inspect specific media
```bash
curl -I http://localhost:8080/app/assets/logo.png
```
* `-I`/`--head`: Tells `curl` to print only the HTTP response headers without any body content  
*Note: To serve additional media assets, place them in the assets/ directory. They will be automatically available at `http://localhost:8080/app/assets/<filename>`.*

### Check if the server is available
To check if the server is up and ready to receive traffic (only accepts `GET` requests):

```bash
curl -i http://localhost:8080/api/healthz
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
...

OK
```

The server will reject any HTTP method other than `GET` at this endpoint:

```bash
curl -i -X POST http://localhost:8080/api/healthz
```
* `-X [METHOD]`: Tells `curl` what HTTP method to use.

Expected response:
```text
HTTP/1.1 405 Method Not Allowed
...
```

### Check how many requests have been served
```bash
curl -i http://localhost:8080/admin/metrics
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: text/html
...

<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 3 times!</p>
  </body>
</html>
```
*Note: the `3` is expected to be any positive integer representing the count of requests served since the server was last started or the request counter was reset.*

### Reset the request counter and user database
Only accepts `POST` requests. Only accessible if `PLATFORM=dev`:

```bash
curl -i -X POST http://localhost:8080/admin/reset
```

Expected response:
```text
Hits reset to 0 and database reset to initial state.
```

*Note: Sending any non-`POST` HTTP request to `/admin/reset` will result in a `405 Method Not Allowed` response.*

### Validate a chirp
Chirps can have a maximum of 140 characters and any profane words (`kerfuffle`, `sharbert`, `fornax`) are automatically replaced with `****`.

#### Valid Chirp
```bash
curl -i -X POST http://localhost:8080/api/validate_chirp \
  -H "Content-Type: application/json" \
  -d '{"body": "This is a valid chirp"}'
```
* `-H`/`--header`: Sets a request header.
* `-d`/`--data`: Sets the request body (the "data"). If used, `curl` will automatically switch to use the `POST` method if one wasn't specified explicitly with `-X`.

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: application/json
...

{"cleaned_body":"This is a valid chirp"}
```

#### Too long
```bash
curl -i -X POST http://localhost:8080/api/validate_chirp \
  -H "Content-Type: application/json" \
  -d '{"body": "lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut pharetra finibus enim eu mattis. Phasellus malesuada nibh at lacus fringilla nam."}'
```

Expected response:
```text
HTTP/1.1 400 Bad Request
Content-Type: application/json
...

{"error":"Chirp is too long"}
```

#### One bad word
```bash
curl -i -X POST http://localhost:8080/api/validate_chirp \
  -H "Content-Type: application/json" \
  -d '{"body": "What a kerfuffle situation"}'
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: application/json
...

{"cleaned_body":"What a **** situation"}
```
*Note: Profanity matching is case-insensitive, so `Kerfuffle`, `KERFUFFLE`, `kerFufFle`, etc. are all replaced. Words are space-separated, so `kerfuffle!`, `kerfuffle,` etc. would **not** be replaced.*

#### Two bad words
```bash
curl -i -X POST http://localhost:8080/api/validate_chirp \
  -H "Content-Type: application/json" \
  -d '{"body": "This sharbert is a really good fornax"}'
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: application/json
...

{"cleaned_body":"This **** is a really good ****"}
```

### Create new user
```bash
curl -i -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

Expected response:
```text
HTTP/1.1 201 Created
Content-Type: application/json
...

{
  "id": "50746277-23c6-4d85-a890-564c0044c2fb",
  "created_at": "2021-07-07T00:00:00Z",
  "updated_at": "2021-07-07T00:00:00Z",
  "email": "user@example.com"
}
```
*Note: The `id` field will be a random UUID. The `created_at` and `updated_at` fields should show around when the command was run.*
