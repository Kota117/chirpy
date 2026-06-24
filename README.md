# Chirpy

A fully-featured social media platform API written in Go.


## Prerequisites
- **Go**: Version 1.22+ installed on the local machine.
- **curl**: For manually testing API endpoints via the terminal.
- **jq**: For manually testing API endpoints via the terminal.
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
- **User Creation**: Creates a new user via `POST /api/users`. Accepts an `email` and `password` in the JSON request body. The password is hashed with Argon2 before storage. Returns the user's `id`, `created_at`, `updated_at`, and `email` (never the hashed password).
- **User Login**: Authenticates a user via `POST /api/login`. Accepts an `email` and `password`. Returns the user resource and a signed JWT token on success, or `401 Unauthorized` with the message "Incorrect email or password" if the email lookup or password comparison fails.
- **Chirp Creation**: Creates a new chirp via `POST /api/chirps`. Validates that the chirp is no longer than 140 characters and replaces profane words (`kerfuffle`, `sharbert`, `fornax`) with `****`. Saves the chirp to the database and returns the full chirp resource with a `201 Created` status. Requires a valid JWT in the `Authorization: Bearer <token>` header. The user ID is derived from the token, not the request body. Returns `401 Unauthorized` if the JWT is missing or invalid.
- **Chirp Retrieval**: Retrieves all chirps stored in the database via `GET /api/chirps`. Returns them as a JSON array sorted in ascending order by `created_at`.
- **Single Chirp Retrieval**: Retrieves a single chirp by its UUID via `GET /api/chirps/{chirpID}`. Returns `404 Not Found` if the chirp does not exist.


## Security Notes

### Passwords are never stored in plain text
User passwords are hashed using [Argon2](https://en.wikipedia.org/wiki/Argon2) (via the `github.com/alexedwards/argon2id` library) before being written to the database. Hashing is a one-way function: even if the database is compromised, the original passwords cannot be recovered from the stored hashes.

### Hashed passwords are never returned in responses
The `User` struct tags the password field with `json:"-"`, ensuring it is excluded from all JSON responses. The API never echoes back password data, hashed or otherwise.

### Login errors are intentionally vague
The `POST /api/login` endpoint returns the same `401 Unauthorized` message, "Incorrect email or password", whether the email doesn't exist OR the password is wrong. This prevents attackers from discovering which emails are registered (an enumeration attack).

### Raw passwords in requests rely on HTTPS
Passwords are sent as plain text in the request body. This is only safe because production traffic is encrypted end-to-end with HTTPS/TLS. Never run this API over plain HTTP in production.

### JWT
JWTs are signed, not encrypted — The payload is Base64-encoded and readable by anyone. Never store sensitive data (like passwords) in a JWT. The signature only guarantees the token hasn't been tampered with.


## Project Structure
```text
.
├── assets/                   # Static assets like images and logos
│   └── logo.png
├── internal/
│   ├── auth/                 # Password hashing & comparison and JWT creation & validation helpers
│   │   ├── auth.go
│   │   └── auth_test.go
│   └── database/             # SQLC-generated Go database code
├── sql/                     
│   ├── queries/              # SQLC query definitions
│   │   ├── chirps.sql
│   │   └── users.sql
│   └── schema/               # Goose migration files
│       ├── 001_users.sql
│       ├── 002_chirps.sql
│       └── 003_password.sql
├── .env                      # Local environment variables (not version-tracked)
├── .gitignore                # Disables version-tracking for any included files/folders
├── go.mod                    # Go module definition
├── handler_chirps_create.go  # Handler for creating and validating a new chirp
├── handler_chirps_get.go     # Handler for retrieving chirps (all or by uuid)
├── handler_login.go          # Handler for authenticating a user
├── handler_metrics.go        # Handler for getting the number of requests since the server was last started
├── handler_readiness.go      # Handler for testing if the server is up and ready to receive traffic
├── handler_reset.go          # Handler for resetting the request counter
├── handler_users_create.go   # Handler for creating a new user
├── index.html                # Root HTML file served at http://localhost:8080
├── json.go                   # Shared helpers for encoding JSON responses and errors
├── main.go                   # Entry point for the Go server
└── sqlc.yaml                 # SQLC configuration
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
- `github.com/alexedwards/argon2id` — Argon2 password hashing wrapper
- `github.com/golang-jwt/jwt/v5` — JWT creation and validation


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
JWT_SECRET="your-generated-secret-here"
```
*Note: Generate a JWT secret with: `openssl rand -base64 64`*

Example (Linux/WSL):
```bash
DB_URL="postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
JWT_SECRET="Rvn5iIEn+4CdTS9u7QEDH5Z6sttc73hsF+jqAKDtL90AY2lMHS5obnk1FL9Lvk75Iqr7fpVxIyXlAj6Km7de9Q=="
```
*Note: This `JWT_SECRET` is just an example of running the suggested command and is not the secret on this machine.*


## Database Migrations
Chirpy uses [Goose](https://github.com/pressly/goose) to manage database schema migrations. Migration files live in `sql/schema/` and are plain `.sql` files with special Goose comments.

### Running Migrations
To apply the latest database schema, `cd` into the `sql/schema` directory, then run:
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
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable" up
```

**Verify the migration applied successfully:**
```bash
psql "<connection_string>"
\dt
```

Example (Linux/WSL):
```bash
psql "postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
\dt
```


## Current Schema
| Table  | Column          | Type      | Constraints                                      |
| ------ | --------------- | --------- | ------------------------------------------------ |
| users  | id              | UUID      | PRIMARY KEY                                      |
| users  | created_at      | TIMESTAMP | NOT NULL                                         |
| users  | updated_at      | TIMESTAMP | NOT NULL                                         |
| users  | email           | TEXT      | NOT NULL, UNIQUE                                 |
| users  | hashed_password | TEXT      | NOT NULL, DEFAULT 'unset'                        |
| chirps | id              | UUID      | PRIMARY KEY                                      |
| chirps | created_at      | TIMESTAMP | NOT NULL                                         |
| chirps | updated_at      | TIMESTAMP | NOT NULL                                         |
| chirps | body            | TEXT      | NOT NULL                                         |
| chirps | user_id         | UUID      | NOT NULL, REFERENCES users(id) ON DELETE CASCADE |


## Generating Database Code (SQLC)
Chirpy uses [SQLC](https://sqlc.dev/) to generate type-safe Go code from SQL queries.

To regenerate the `internal/database` package after modifying queries in `sql/queries/`:
```bash
sqlc generate
```
*Note: This command should be run from the `root` of the project.*

## Usage
| Endpoint                | Method | Description                                                   |
| ----------------------- | ------ | ------------------------------------------------------------- |
| `/app/*`                | GET    | Serves static frontend files                                  |
| `/api/healthz`          | GET    | Readiness check                                               |
| `/api/users`            | POST   | Create new user                                               |
| `/api/login`            | POST   | Authenticate a user with email and password                   |
| `/api/chirps`           | POST   | Create a new chirp (max 140 chars, profanity filtered)        |
| `/api/chirps`           | GET    | Retrieve all chirps (sorted ascending by creation time)       |
| `/api/chirps/{chirpID}` | GET    | Retrieve a single chirp by ID (returns 404 if not found)      |
| `/admin/metrics`        | GET    | Retrieve hit counter (HTML)                                   |
| `/admin/reset`          | POST   | Reset hit counter and delete all users (dev environment only) |


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

The terminal window running the server will show a message like this:
```text
2026/06/21 10:01:42 Serving files from . on port: 8080

```


## Manually Testing the Server
While the server is running in one terminal window, the back-end can be manually tested in another terminal window. Additionally, the front-end content can be viewed in a browser at `http://localhost:8080/app/`.  
  
Admin metrics can be viewed at `http://localhost:8080/admin/metrics`.

### Check if the server is available
Requires `curl` to be installed: `sudo apt install curl`. To check if the server is up and ready to receive traffic (only accepts `GET` requests):

```bash
curl -i http://localhost:8080/api/healthz
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Wed, 24 Jun 2026 22:01:25 GMT
Content-Length: 3

OK
```
*Note: The `Date` shows from when the request originated.*

The server will reject any HTTP method other than `GET` at this endpoint:

```bash
curl -i -X POST http://localhost:8080/api/healthz
```
* `-X [METHOD]`: Tells `curl` what HTTP method to use.

Expected response:
```text
HTTP/1.1 405 Method Not Allowed
Allow: GET, HEAD
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 24 Jun 2026 22:02:48 GMT
Content-Length: 19

Method Not Allowed
```

### Inspect the contents of index.html
```bash
curl -i http://localhost:8080/app/
```
* `-i`/`--include`: Tells `curl` to print the HTTP response headers along with any body content.
*Note: The default HTTP method used by curl is `GET`.*

Expected response:
```text
HTTP/1.1 200 OK
Accept-Ranges: bytes
Content-Length: 65
Content-Type: text/html; charset=utf-8
Last-Modified: Tue, 16 Jun 2026 16:19:19 GMT
Date: Wed, 24 Jun 2026 22:03:34 GMT

<html>
  <body>
    <h1>Welcome to Chirpy</h1>
  </body>
</html>
```

### Inspect specific media

#### Existing media
```bash
curl -I http://localhost:8080/app/assets/logo.png
```
* `-I`/`--head`: Tells `curl` to print only the HTTP response headers without any body content.  
*Note: To serve additional media assets, place them in the assets/ directory. They will be automatically available at `http://localhost:8080/app/assets/<filename>`.*

Expected response:
```text
HTTP/1.1 200 OK
Accept-Ranges: bytes
Content-Length: 32010
Content-Type: image/png
Last-Modified: Tue, 16 Jun 2026 16:40:30 GMT
Date: Wed, 24 Jun 2026 22:04:00 GMT

```

#### Non-existing media
```bash
curl -I http://localhost:8080/app/assets/fake_logo.png
```

Expected response:
```text
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 24 Jun 2026 22:04:17 GMT
Content-Length: 19

```

### Check how many requests have been served
Metrics can also be viewed in a web browser at `http://localhost:8080/admin/metrics`.
```bash
curl -i http://localhost:8080/admin/metrics
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: text/html
Date: Wed, 24 Jun 2026 22:04:35 GMT
Content-Length: 114


<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 3 times!</p>
  </body>
</html>
```
*Note: the `3` is expected to be any positive integer representing the count of requests served since the server was last started or the request counter was reset.*

### Reset the request counter and user database
Only accepts `POST` requests. Only accessible if `PLATFORM=dev`, otherwise returns `403 Forbidden`.

```bash
curl -i -X POST http://localhost:8080/admin/reset
```

Expected response:
```text
HTTP/1.1 200 OK
Date: Wed, 24 Jun 2026 22:04:58 GMT
Content-Length: 53
Content-Type: text/plain; charset=utf-8

Hits reset to 0 and database reset to initial state.
```
*Note: Sending any non-`POST` HTTP request to `/admin/reset` will result in a `405 Method Not Allowed` response.*

### Create new user
It is helpful to always reset the database before testing so that any previous tests won't affect the current test.
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -i -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}'
```
* `-s`/`--silent`: Suppresses the progress meter.
* `> /dev/null`: Discards the response body.
* `-H`/`--header`: Sets a request header.
* `-d`/`--data`: Sets the request body (the "data"). If used, `curl` will automatically switch to use the `POST` method if one wasn't specified explicitly with `-X`.

Expected response:
```text
HTTP/1.1 201 Created
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:05:19 GMT
Content-Length: 158

{
  "id":"8dc99490-ce27-4e16-b0ca-59c96d5388b6",
  "created_at":"2026-06-24T16:05:19.421934Z",
  "updated_at":"2026-06-24T16:05:19.421934Z",
  "email":"user@example.com"
}
```
*Note: The `id` field will be a random UUID. The `created_at` and `updated_at` fields should show around when the command was run (in local time).*
*Note: The returned JSON will likely be collapsed. The response is prettified here for clarity.*

### Login

#### Correct password
First reset the database, create a user, then log in with the same credentials.
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

curl -i -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}'
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:06:13 GMT
Content-Length: 386

{
  "id":"55defd5f-2642-4c84-b621-dd777a7b405d",
  "created_at":"2026-06-24T16:06:13.886916Z",
  "updated_at":"2026-06-24T16:06:13.886916Z",
  "email":"user@example.com",
  "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJjaGlycHktYWNjZXNzIiwic3ViIjoiNTVkZWZkNWYtMjY0Mi00Yzg0LWI2MjEtZGQ3NzdhN2I0MDVkIiwiZXhwIjoxNzgyMzQyMzczLCJpYXQiOjE3ODIzMzg3NzN9.JuOQnG39LzzxKbrmGBRrcSj6eWCe6f1Vjxn8i-yUxIk"
}
```

#### Incorrect password
First reset the database, create a user, then log in with the wrong password credentials.
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

curl -i -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "12345"}'
```

Expected response:
```text
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:07:36 GMT
Content-Length: 39

{"error":"Incorrect email or password"}
```

### Create a chirp
Chirps can have a maximum of 140 characters and any profane words (`kerfuffle`, `sharbert`, `fornax`) are automatically replaced with `****`. The user is identified by the JWT, so the user must exist and be logged in.

#### Valid chirp
Requires `jq` to be installed: `sudo apt install jq`. First resets the database, then creates a `user`, then logs into that user, and then creates the `chirp`; this ensures the `user` exists and there is a valid `token` used.
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' | jq -r '.token')

curl -i -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "Hello, World!"}'
```

Expected response:
```text
HTTP/1.1 201 Created
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:12:09 GMT
Content-Length: 203

{
  "id":"dba19cca-3f84-4b0a-ae97-aefad1e54d67",
  "created_at":"2026-06-24T16:12:09.058969Z",
  "updated_at":"2026-06-24T16:12:09.058969Z",
  "body":"Hello, World!",
  "user_id":"5dc8097c-27af-4e07-8486-cd3a2cbcfef5"
}
```
*Note: The `user_id` field will be a random UUID correlated with the currently logged in user.*

#### Invalid token
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

TOKEN="not a valid token"

curl -i -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "Hello, World!"}'
```

Expected response:
```text
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:27:39 GMT
Content-Length: 33

{"error":"Couldn't validate JWT"}
```

#### Too long chirp
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' | jq -r '.token')

curl -i -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut pharetra finibus enim eu mattis. Phasellus malesuada nibh at lacus fringilla nam."}'
```

Expected response:
```text
HTTP/1.1 400 Bad Request
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:15:08 GMT
Content-Length: 29

{"error":"Chirp is too long"}
```

#### One bad word in chirp
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' | jq -r '.token')

curl -i -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "What a kerfuffle situation"}'
```

Expected response:
```text
HTTP/1.1 201 Created
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:15:43 GMT
Content-Length: 211

{
  "id":"edf06cdf-0928-4670-8745-ab51607daac4",
  "created_at":"2026-06-24T16:15:43.779179Z",
  "updated_at":"2026-06-24T16:15:43.779179Z",
  "body":"What a **** situation",
  "user_id":"3372c75f-e78f-40e7-8f3c-87a234897f6b"
}
```
*Note: Profanity matching is case-insensitive, so `Kerfuffle`, `KERFUFFLE`, `kerFufFle`, etc. are all replaced. Words are space-separated, so `kerfuffle!`, `kerfuffle,` etc. would **not** be replaced.*

#### Multiple bad words in chirp
```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' | jq -r '.token')

curl -i -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "What in the sharbert is fornax"}'
```

Expected response:
```text
HTTP/1.1 201 Created
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:17:04 GMT
Content-Length: 212

{
  "id":"785cd563-07c3-43e6-bef6-2296181c5211",
  "created_at":"2026-06-24T16:17:04.11606Z",
  "updated_at":"2026-06-24T16:17:04.11606Z",
  "body":"What in the **** is ****",
  "user_id":"3b951e56-f159-4f00-9863-a40a8c585532"
}
```

### Retrieve all chirps
Retrieves a list of all chirps in the database, ordered chronologically. First resets the database, creates a user, creates two chirps, and then retrieves all chirps.

```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' | jq -r '.token')

curl -s -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "Hello, World!"}' > /dev/null

curl -s -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "This is another chirp!"}' > /dev/null

curl -i http://localhost:8080/api/chirps
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:18:52 GMT
Content-Length: 416

[
  {
    "id":"978e160a-0a20-4a53-9fcc-9538b92c2565",
    "created_at":"2026-06-24T16:18:52.023194Z",
    "updated_at":"2026-06-24T16:18:52.023194Z",
    "body":"Hello, World!",
    "user_id":"bc5b80e2-f3d5-40ef-82ea-bdf641c3197c"
  },
  {
    "id":"1aaf0a09-5a5a-4f87-bef8-57ef4168e353",
    "created_at":"2026-06-24T16:18:52.03205Z",
    "updated_at":"2026-06-24T16:18:52.03205Z",
    "body":"This is another chirp!",
    "user_id":"bc5b80e2-f3d5-40ef-82ea-bdf641c3197c"
  }
]
```
*Note: Notice that both chirps' `user_id` are the same because the same user posted each one using one unique token.*

### Retrieve a single chirp

#### Retrieving a valid chirp
First resets the database, creates a user, creates some chirps, and then uses a generated chirp's ID to retrieve only that specific chirp.

```bash
curl -s -X POST http://localhost:8080/admin/reset > /dev/null

curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "1234"}' | jq -r '.token')

curl -s -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "Hello, World!"}' > /dev/null

CHIRP_ID=$(curl -s -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "This is a single targeted chirp!"}' | jq -r '.id')

curl -s -X POST http://localhost:8080/api/chirps \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"body": "This is a more recent chirp than the target!"}' > /dev/null

curl -i http://localhost:8080/api/chirps/$CHIRP_ID
```

Expected response:
```text
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:24:08 GMT
Content-Length: 222

{
  "id":"8f021620-3143-4fc3-848a-5b87169e785f",
  "created_at":"2026-06-24T16:24:08.542935Z",
  "updated_at":"2026-06-24T16:24:08.542935Z",
  "body":"This is a single targeted chirp!",
  "user_id":"582f2b41-ceb2-4586-84f0-f670592c3917"
}
```

#### Retrieving a non-existing chirp ID
```bash
curl -i http://localhost:8080/api/chirps/00000000-0000-0000-0000-000000000000
```

Expected response:
```text
HTTP/1.1 404 Not Found
Content-Type: application/json
Date: Wed, 24 Jun 2026 22:25:04 GMT
Content-Length: 30

{"error":"Couldn't get chirp"}
```
