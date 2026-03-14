# Food Delivery - Microservices Platform

A production-ready food delivery backend built with Go microservices, featuring event-driven architecture, JWT authentication, API gateway, service discovery, and full observability stack.

## Architecture Overview

```
Client
  │
  ▼
KrakenD API Gateway (port 8080)
  │  ├── JWT Validation (RS256)
  │  ├── Correlation ID Injection
  │  └── Prometheus Metrics (port 9091)
  │
  ├──► Auth Service (port 3005)
  │      └── PostgreSQL (auth_db)
  │
  ├──► Order Service (port 3000)
  │      ├── PostgreSQL (order_db)
  │      ├── RabbitMQ Publisher ──────────────────────────────┐
  │      ├── Circuit Breaker (Sony GoBreaker)                  │
  │      └── Consul Client                                     │
  │                                                            ▼
  └──► Kitchen Service (port 3001)                    RabbitMQ (port 5672)
         ├── PostgreSQL (kitchen_db)                      │
         ├── Consul Registration                          │
         └── Kitchen Worker ◄────────────────────────────┘
               └── Consumes order.created events

Observability:
  Prometheus (9090) ◄── scrapes all services
  Grafana (3002) ◄── Prometheus + Loki dashboards
  Loki (3100) ◄── Promtail ◄── Docker logs
  Consul UI (8500) ◄── service registry
  RabbitMQ UI (15672) ◄── message broker management
```

## Services

| Service | Port | Responsibility |
|---------|------|----------------|
| API Gateway (KrakenD) | 8080 | Routing, JWT validation, rate limiting |
| Auth Service | 3005 | User registration, login, JWT issuance |
| Order Service | 3000 | Order CRUD, publishes order events |
| Kitchen Service | 3001 | Kitchen ticket management, consumes order events |
| Kitchen Worker | — | RabbitMQ consumer (separate process) |

## Tech Stack

- **Language:** Go 1.25.7
- **Web Framework:** Go Fiber v2
- **ORM:** GORM with PostgreSQL 15
- **Message Broker:** RabbitMQ 3
- **API Gateway:** KrakenD
- **Service Discovery:** HashiCorp Consul
- **Auth:** JWT RS256 (RSA 2048-bit key pair)
- **Metrics:** Prometheus + Grafana
- **Logs:** Zerolog → Promtail → Loki → Grafana
- **Resilience:** Sony GoBreaker (circuit breaker)
- **Containers:** Docker + Docker Compose

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- Go 1.24+ (for local development only)

### Run with Docker Compose

```bash
# Clone the repository
git clone <repo-url>
cd food-delivery

# Start all services
docker-compose up -d --build

# Check all containers are running
docker-compose ps
```

### Service URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| API Gateway | http://localhost:8080 | — |
| Grafana | http://localhost:3002 | admin / admin |
| Prometheus | http://localhost:9090 | — |
| RabbitMQ UI | http://localhost:15672 | guest / guest |
| Consul UI | http://localhost:8500 | — |
| Order Swagger | http://localhost:3000/swagger/index.html | — |

## API Reference

All requests go through the API Gateway at `http://localhost:8080`.

### Authentication

#### Register
```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "password": "secret", "email": "john@example.com"}'
```

#### Login
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "password": "secret"}'
# Response: {"token": "eyJhbGc..."}
```

### Orders (requires JWT)

#### Create Order
```bash
curl -X POST http://localhost:8080/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"customer_id": "cust_001", "total_amount": 45.50}'
```

### Kitchen

#### Check Kitchen Status
```bash
curl http://localhost:8080/v1/kitchen/status/{orderId}
```

## Order Flow

```
sequenceDiagram
  Client → Gateway: POST /v1/orders + JWT
  Gateway → Gateway: Validate RS256 JWT, extract user_id
  Gateway → OrderService: POST /api/v1/orders + X-User-Id header
  OrderService → PostgreSQL: Save order (status: Pending)
  OrderService → RabbitMQ: Publish order.created event
  OrderService → Client: 201 Created
  RabbitMQ → KitchenWorker: Deliver event
  KitchenWorker → PostgreSQL: Create KitchenTicket (status: Received)
```

## Event-Driven Architecture

### RabbitMQ Configuration

| Parameter | Value |
|-----------|-------|
| Exchange | `order_events` (topic) |
| Queue | `kitchen_order_queue` |
| Routing Key | `order.created` |
| Durability | Durable (survives restarts) |

### Message Payload
```json
{
  "order_id": 1,
  "items": "[]"
}
```

Correlation ID is passed via RabbitMQ message headers for end-to-end request tracing.

## Authentication & Security

- **Password hashing:** bcrypt (cost 10)
- **JWT algorithm:** RS256 (RSA 2048-bit)
- **Token expiry:** 24 hours
- **Token validation:** Performed at gateway level before requests reach backend services
- **Key rotation:** Replace `private_key.pem`, `public_key.pem`, and `public_key.json` (JWKS format)

## Environment Variables

### Auth Service
```env
PORT=3002
DB_URL=postgres://admin:admin@db:5432/auth_db?sslmode=disable
PRIVATE_KEY_PATH=/app/private_key.pem
```

### Order Service
```env
PORT=3000
DB_URL=postgres://admin:admin@db:5432/order_db?sslmode=disable
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
CONSUL_ADDRESS=consul:8500
```

### Kitchen Service
```env
PORT=3001
DB_URL=postgres://admin:admin@db:5432/kitchen_db?sslmode=disable
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
CONSUL_ADDRESS=consul:8500
SERVICE_ADDRESS=kitchen-api
```

## Resilience Patterns

### Circuit Breaker (Order → Kitchen)
| Parameter | Value |
|-----------|-------|
| Max requests (half-open) | 3 |
| Interval | 5 seconds |
| Timeout (open state) | 30 seconds |
| Trip condition | 3 consecutive failures |

When the circuit is open, Order Service returns: `"Kitchen service is unavailable"` immediately without waiting.

## Observability

### Prometheus Targets
- Order Service: `order-service-api:3000/metrics`
- Kitchen Service: `kitchen-service-api:3001/metrics`
- Auth Service: `auth-service-api:3002/metrics`
- Gateway: `api-gateway:9091/metrics`

### Structured Logging
All services use **Zerolog** for structured JSON logs. Every log entry includes:
- `correlation_id` — end-to-end request trace ID
- `service` — service name
- `method`, `path`, `status`, `latency`

### Grafana Dashboards
1. **Application Metrics** — HTTP request rates, latencies, error rates
2. **Go Runtime** — memory, goroutines, GC activity
3. **Loki Logs** — centralized log search and filtering

## Project Structure

```
food-delivery/
├── auth-service/
│   ├── cmd/main.go
│   └── internal/
│       ├── handler/        # HTTP handlers
│       ├── middleware/      # Logger
│       ├── model/           # User model
│       ├── repository/      # Database layer
│       └── service/         # JWT generation, business logic
├── order-service/
│   ├── cmd/main.go
│   ├── docs/                # Swagger/OpenAPI docs
│   └── internal/
│       ├── handler/         # Order CRUD handlers
│       ├── middleware/      # Correlation ID, logger
│       ├── model/           # Order model
│       ├── repository/      # Database layer
│       └── service/         # RabbitMQ publish, circuit breaker
├── kitchen-service/
│   ├── cmd/
│   │   ├── main.go          # API server
│   │   └── worker/main.go   # RabbitMQ consumer
│   └── internal/
│       ├── handler/         # Ticket handlers
│       ├── middleware/      # Logger
│       ├── model/           # KitchenTicket model
│       ├── repository/      # Database layer
│       ├── service/         # Kitchen business logic
│       └── worker/          # Event consumer
├── gateway/
│   └── plugin/              # Correlation ID injector plugin
├── krakend.json             # API Gateway routing & JWT config
├── docker-compose.yaml      # Full stack orchestration
├── prometheus.yml           # Metrics scrape config
├── promtail-config.yaml     # Log shipping config
├── init.sql                 # Creates order_db, kitchen_db, auth_db
├── private_key.pem          # RSA private key (JWT signing)
├── public_key.pem           # RSA public key
└── public_key.json          # JWKS format (KrakenD JWT validation)
```

## Database Schema

Tables are auto-migrated by GORM on service startup.

**users** (auth_db)
```
id, username (unique), password (bcrypt), email (unique), created_at, updated_at
```

**orders** (order_db)
```
id, customer_id, total_amount (decimal 10,2), status (default: Pending), created_at, updated_at
```

**kitchen_tickets** (kitchen_db)
```
id, order_id (unique), items (JSON), status (Received|Cooking|Ready), created_at
```

## Development

### Regenerate Swagger Docs (Order Service)
```bash
cd order-service
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -d .
docker-compose up -d --build api
```

### Database Scripts
```bash
# Setup tables
go run scripts/setup_db.go

# Seed sample data
go run scripts/seed_db.go
```

### Rebuild a Single Service
```bash
docker-compose up -d --build auth-service
docker-compose up -d --build api           # order-service
docker-compose up -d --build kitchen-api
```
