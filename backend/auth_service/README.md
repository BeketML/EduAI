# 🔐 Auth Service

**Auth Service** is a Go-based authentication microservice, part of the **AI-Powered Educational Platform for Lectures** project. It provides user registration, login, JWT-based authentication (Access + Refresh tokens), and user profile retrieval. The service is designed following Clean Architecture (DDD-style) principles and is ready for deployment with Docker.

---

## 📖 Table of Contents

1. [Project Overview](#-project-overview)
2. [Architecture](#-architecture)
3. [Folder Structure](#-folder-structure)
4. [Technologies Used](#-technologies-used)
5. [Request ID Middleware](#-request-id-middleware--distributed-tracing)
6. [Environment Variables](#-environment-variables)
7. [Running Locally (without Docker)](#-running-locally-without-docker)
8. [Running with Docker Compose](#-running-with-docker-compose)
9. [API Endpoints](#-api-endpoints)
10. [Example Requests](#-example-requests)
11. [X-Request-ID Usage](#-x-request-id-usage)
12. [Inter-Service JWT Validation via /me](#-inter-service-jwt-validation-via-me)
13. [Database Migrations](#-database-migrations)

---

## 🧩 Project Overview

The Auth Service handles all authentication concerns for the platform:

- **Register** — create a new user account with hashed password (bcrypt)
- **Login** — authenticate with username/password, receive Access + Refresh JWT tokens
- **Refresh** — obtain a new token pair using a valid refresh token
- **Logout** — invalidate the refresh token stored in the database
- **Me** — retrieve the authenticated user's profile from an access token

Other microservices (Content Service, AI Service) can validate user identity by calling the `/api/v1/auth/me` endpoint with the user's access token.

---

## 🏗 Architecture

The service follows **Clean Architecture** (DDD-style), with clearly separated layers:

```
┌──────────────────────────────────────────────┐
│                 Interfaces                   │
│         (HTTP Handlers, Middleware)           │
├──────────────────────────────────────────────┤
│                  Usecase                     │
│      (Business Logic — AuthService)          │
├──────────────────────────────────────────────┤
│               Domain                         │
│        (Entities, Repository Interfaces)     │
├──────────────────────────────────────────────┤
│             Infrastructure                   │
│  (Postgres Repos, JWT Manager, Logger, DB)   │
└──────────────────────────────────────────────┘
```

| Layer            | Responsibility                                                        |
|------------------|-----------------------------------------------------------------------|
| **Domain**       | `User` entity definition; repository interface contracts              |
| **Usecase**      | `AuthService` — Register, Login, Refresh, Logout, Me                  |
| **Interfaces**   | Gin HTTP handlers, JWT auth middleware, RequestID middleware           |
| **Infrastructure** | PostgreSQL repositories, JWT token manager, structured logger, DB connection |

---

## 📁 Folder Structure

```
auth_service/
├── app/
│   └── cmd/
│       └── main.go                  # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go                # Config struct definition
│   ├── domain/
│   │   └── user.go                  # User entity (domain model)
│   ├── infrastructure/
│   │   ├── auth/
│   │   │   └── jwt.go               # JWT token manager (access + refresh)
│   │   ├── logger/
│   │   │   └── slog.go              # Structured logger with request_id support
│   │   ├── postgres/
│   │   │   ├── connection.go        # Database connection helper
│   │   │   ├── retry.go             # Connection retry logic
│   │   │   └── user/
│   │   │       └── auth.go          # Postgres repository implementation
│   │   └── repository/
│   │       └── repository.go        # Repository interface aggregation
│   ├── interfaces/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── auth.go          # Auth HTTP handlers (register, login, etc.)
│   │       │   ├── handler.go       # Handler struct & router initialization
│   │       │   ├── middleware.go     # JWT auth middleware (userIdentity)
│   │       │   ├── response.go      # Response DTOs
│   │       │   └── router.go        # HTTP server wrapper
│   │       └── middleware/
│   │           └── request_id.go    # X-Request-ID middleware
│   └── usecase/
│       ├── service.go               # Service interface & constructor
│       └── auth/
│           └── auth.go              # Auth business logic implementation
├── migrations/
│   ├── 20260222155205_create_users_table.up.sql
│   └── 20260222155205_create_users_table.down.sql
├── config.yml                       # Application config (port, DB settings)
├── docker-compose.yml               # Docker Compose for service + DB + migrations
├── Dockerfile                       # Multi-stage Docker build
├── go.mod
└── go.sum
```

---

## 🛠 Technologies Used

| Technology                | Purpose                                      |
|---------------------------|----------------------------------------------|
| **Go 1.24**               | Primary language                             |
| **Gin**                   | HTTP web framework                           |
| **PostgreSQL 15**         | Relational database                          |
| **sqlx**                  | SQL extensions for Go (named queries, struct scanning) |
| **golang-jwt/jwt/v5**     | JWT token generation & validation            |
| **bcrypt** (`golang.org/x/crypto`) | Password hashing                    |
| **google/uuid**           | UUID generation                              |
| **spf13/viper**           | Configuration management (YAML + env)        |
| **slog** (stdlib)         | Structured logging with request_id context   |
| **Docker & Docker Compose** | Containerized deployment                   |
| **golang-migrate/migrate** | Database schema migrations                  |
| **swaggo/gin-swagger**    | Swagger API documentation                   |
| **gin-contrib/cors**      | CORS middleware                               |

---

## 🆔 Request ID Middleware & Distributed Tracing

Every incoming HTTP request passes through a `RequestID` middleware that:

1. Checks for an existing `X-Request-ID` header
2. Generates a new UUID if the header is missing
3. Stores the `request_id` in the request's `context.Context`
4. Sets the `X-Request-ID` response header
5. Makes `request_id` available for logging across **all layers**

```go
// middleware/request_id.go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        reqID := r.Header.Get("X-Request-ID")
        if reqID == "" {
            reqID = uuid.NewString()
        }
        ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
        w.Header().Set("X-Request-ID", reqID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

The logger automatically extracts `request_id` from context and includes it in every log entry:

```
time=2026-02-23T10:00:00Z level=INFO msg="creating user successfully" request_id=a1b2c3d4-...
```

---

## ⚙️ Environment Variables

Create a `.env` file in the `auth_service/` directory:

```env
# Database
DB_PASSWORD=postgres

# JWT Secrets
JWT_ACCESS_SECRET=your-access-secret-key
JWT_REFRESH_SECRET=your-refresh-secret-key
```

Additional configuration is managed via `config.yml`:

```yaml
port: "8080"

db:
  username: "postgres"
  host: "auth-db"        # Use "localhost" when running without Docker
  port: 5432
  dbname: "auth_db"
  sslmode: "disable"
```

---

## 🚀 Running Locally (without Docker)

### Prerequisites

- Go 1.24+
- PostgreSQL 15+
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Steps

1. **Clone the repository:**

   ```bash
   git clone https://github.com/your-org/AI-Powered-Educational-Platform-for-Lectures.git
   cd AI-Powered-Educational-Platform-for-Lectures/auth_service
   ```

2. **Create and configure PostgreSQL database:**

   ```sql
   CREATE DATABASE auth_db;
   ```

3. **Run migrations:**

   ```bash
   migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable" up
   ```

4. **Update `config.yml`** — set `db.host` to `localhost`:

   ```yaml
   db:
     host: "localhost"
   ```

5. **Set environment variables:**

   ```bash
   # Linux / macOS
   export DB_PASSWORD=postgres
   export JWT_ACCESS_SECRET=my-access-secret
   export JWT_REFRESH_SECRET=my-refresh-secret

   # Windows (cmd)
   set DB_PASSWORD=postgres
   set JWT_ACCESS_SECRET=my-access-secret
   set JWT_REFRESH_SECRET=my-refresh-secret
   ```

6. **Run the service:**

   ```bash
   go run ./app/cmd/main.go
   ```

   The server will start on `http://localhost:8080`.

---

## 🐳 Running with Docker Compose

```bash
cd auth_service
docker-compose up --build
```

This starts three containers:

| Container        | Description                           | Port  |
|------------------|---------------------------------------|-------|
| `auth-db`        | PostgreSQL 15 database                | 5432  |
| `auth-migrate`   | Runs database migrations on startup   | —     |
| `auth_service`   | Auth microservice (Go)                | 8080  |

All containers are connected via the `beket-net` Docker network.

To stop:

```bash
docker-compose down
```

To stop and remove volumes (⚠️ deletes DB data):

```bash
docker-compose down -v
```

---

## 📡 API Endpoints

All endpoints are prefixed with `/api/v1/auth`.

| Method | Endpoint     | Auth Required | Description                              |
|--------|-------------|---------------|------------------------------------------|
| POST   | `/register` | ❌            | Register a new user                      |
| POST   | `/login`    | ❌            | Login and receive JWT tokens             |
| POST   | `/refresh`  | ❌            | Refresh access token using refresh token |
| POST   | `/logout`   | ✅ Bearer     | Logout (invalidate refresh token)        |
| GET    | `/me`       | ✅ Bearer     | Get current authenticated user's profile |

### Swagger Documentation

When running, Swagger UI is available at:

```
http://localhost:8080/swagger/index.html
```

---

## 📝 Example Requests

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Response** `201 Created`:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

**Response** `200 OK`:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Refresh Tokens

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Response** `200 OK`:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <access_token>"
```

**Response** `200 OK`:
```json
{
  "message": "logged out"
}
```

### Get Current User (Me)

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

**Response** `200 OK`:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe"
}
```

---

## 🔗 X-Request-ID Usage

The `X-Request-ID` header enables **distributed tracing** across the microservices architecture.

### How it works

1. A client (or API Gateway) sends a request with an `X-Request-ID` header
2. If not provided, the Auth Service generates one automatically
3. The `request_id` is propagated through the context across all layers:
   - **Handler** → **Usecase** → **Repository** → **Database**
4. Every log entry includes the `request_id` field
5. The `X-Request-ID` is returned in the response headers

### Cross-service tracing

When one service calls another, it should forward the `X-Request-ID`:

```
Client → Auth Service (request_id: abc-123)
       → Content Service (X-Request-ID: abc-123)
       → AI Service (X-Request-ID: abc-123)
```

This allows correlating logs from different services for a single user request:

```
# Auth Service logs
level=INFO msg="creating user successfully" request_id=abc-123

# Content Service logs
level=INFO msg="fetching lectures" request_id=abc-123

# AI Service logs
level=INFO msg="generating summary" request_id=abc-123
```

### Forwarding example (from another Go service)

```go
req, _ := http.NewRequest("GET", "http://auth-service:8080/api/v1/auth/me", nil)
req.Header.Set("Authorization", "Bearer "+accessToken)
req.Header.Set("X-Request-ID", requestIDFromContext)
```

---

## 🔑 Inter-Service JWT Validation via /me

Other microservices (Content Service, AI Service) **do not need to know JWT secrets**. Instead, they validate user identity by calling the Auth Service's `/me` endpoint.

### Flow

```
┌─────────────┐         ┌──────────────┐         ┌──────────────┐
│   Client    │──────▶  │ Content Svc  │──────▶  │ Auth Service │
│             │  JWT    │              │  GET /me │              │
│             │  token  │  Validates   │  + JWT   │  Returns     │
│             │         │  user via    │  token   │  user data   │
│             │         │  Auth Svc    │          │  or 401      │
└─────────────┘         └──────────────┘         └──────────────┘
```

### Example: Validating JWT from Content Service

```go
func validateUser(accessToken, requestID string) (*UserInfo, error) {
    req, _ := http.NewRequest("GET", "http://auth-service:8080/api/v1/auth/me", nil)
    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("X-Request-ID", requestID) // propagate tracing

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusUnauthorized {
        return nil, errors.New("invalid or expired token")
    }

    var user UserInfo
    json.NewDecoder(resp.Body).Decode(&user)
    return &user, nil
}
```

**Response from `/me`:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe"
}
```

If the token is invalid or expired, the Auth Service returns `401 Unauthorized`.

---

## 🗃 Database Migrations

Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate) and are located in the `migrations/` directory.

### Users table schema

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    refresh_token TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Running migrations manually

```bash
# Apply all migrations
migrate -path ./migrations \
  -database "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable" up

# Rollback last migration
migrate -path ./migrations \
  -database "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable" down 1
```

### With Docker Compose

Migrations run automatically on startup via the `auth-migrate` container defined in `docker-compose.yml`. No manual action required.

---

## 📄 License

This project is licensed under the **MIT License**. See `LICENSE` for details.

