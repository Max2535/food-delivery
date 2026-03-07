# Order Service

This is the Order Service for the Food Delivery application, built with [Go Fiber](https://gofiber.io/) and [GORM](https://gorm.io/) with PostgreSQL.

## Features

- RESTful API using Go Fiber v2
- PostgreSQL database integration using GORM
- Docker and Docker Compose support
- OpenAPI/Swagger API documentation

## Prerequisites

- [Go](https://go.dev/doc/install) 1.24+
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose

## Getting Started

### 1. Clone the repository

Ensure you are in the project root:
```bash
cd d:\go-lang\food-delivery\order-service
```

### 2. Environment Variables

The application relies on a `.env` file for configuration. A basic `.env` looks like this:
```env
PORT=3000
DB_URL=postgres://admin:admin@localhost:5432/order_db?sslmode=disable
```
*(Note: In docker-compose, the DB host is overridden to `db`)*

### 3. Run with Docker Compose

To start both the PostgreSQL database and the API service in containers:
```bash
docker-compose up -d --build
```

The API will be available at `http://localhost:3000`.

### 4. Run Locally (Without Docker API)

If you prefer to run the API locally on your machine while keeping the DB in Docker:

1. Start only the database:
   ```bash
   docker-compose up -d db
   ```
2. Download dependencies:
   ```bash
   go mod download
   ```
3. Run the application:
   ```bash
   go run cmd/main.go
   ```

## Utilities and Scripts

The project includes setup and seeding scripts:
- **Setup DB**: `go run scripts/setup_db.go`
- **Seed DB**: `go run scripts/seed_db.go`

## API Documentation (Swagger)

Swagger OpenAPI documentation is integrated. Make sure the application is running, then visit:

[http://localhost:3000/swagger/index.html](http://localhost:3000/swagger/index.html)

### Updating Swagger Docs

Whenever you modify handler annotations, regenerate the swagger docs by running:
```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -d .
```
*(If running in Docker, remember to rebuild the api container `docker-compose up -d --build api` to reflect the updated docs.)*
