# CLAUDE.md — Food Delivery Microservices

Development guide for Claude Code when working in this repository.

## Project Overview

Microservices-based food delivery backend written in Go. Three main services (auth, order, kitchen) communicate through RabbitMQ events and are exposed via KrakenD API gateway. Full observability via Prometheus + Grafana + Loki.

## Repository Structure

```
food-delivery/
├── auth-service/        # JWT issuance, user management
├── order-service/       # Order CRUD + RabbitMQ publisher
├── kitchen-service/     # Kitchen tickets + RabbitMQ consumer worker
├── gateway/             # KrakenD plugin (correlation ID injector)
├── docker-compose.yaml  # Full stack (12 containers)
├── krakend.json         # Gateway routing + JWT validation config
├── prometheus.yml       # Metrics scrape config
├── promtail-config.yaml # Loki log shipping
├── init.sql             # Creates 3 PostgreSQL databases
├── private_key.pem      # RSA-2048 private key (JWT signing)
├── public_key.pem       # RSA public key
└── public_key.json      # JWKS format for KrakenD
```

Each service is a standalone Go module with its own `go.mod`.

## Tech Stack

- **Go 1.25.7**, **Go Fiber v2**, **GORM**
- **PostgreSQL 15** — one DB per service (`order_db`, `kitchen_db`, `auth_db`)
- **RabbitMQ 3** — async event bus between Order and Kitchen
- **KrakenD** — API gateway with JWT RS256 validation
- **Consul** — service discovery (Kitchen registers; Order can discover)
- **Prometheus + Grafana + Loki + Promtail** — observability
- **Zerolog** — structured JSON logging
- **Sony GoBreaker** — circuit breaker in Order Service
- **golang-jwt/jwt/v5** — JWT generation (Auth Service)

## Running the Stack

```bash
# Start everything
docker-compose up -d --build

# Rebuild one service
docker-compose up -d --build api          # order-service
docker-compose up -d --build auth-service
docker-compose up -d --build kitchen-api
docker-compose up -d --build kitchen-worker
```

## Service Ports

| Service | Internal Port | Exposed Port |
|---------|--------------|--------------|
| API Gateway | 8080 | 8080 |
| Order Service | 3000 | 3000 |
| Kitchen Service | 3001 | 3001 |
| Auth Service | 3002 | 3005 |
| PostgreSQL | 5432 | 5432 |
| RabbitMQ | 5672 | 5672 |
| RabbitMQ UI | 15672 | 15672 |
| Consul | 8500 | 8500 |
| Prometheus | 9090 | 9090 |
| Grafana | 3000 | 3002 |
| Loki | 3100 | 3100 |

## Key Architectural Decisions

### Each Service Has Its Own Database
Never share a database between services. Each service owns its schema and connects to its own DB (`order_db`, `kitchen_db`, `auth_db`).

### Authentication Flow
1. Auth Service issues RS256 JWT signed with `private_key.pem`
2. KrakenD validates all protected requests using `public_key.json` (JWKS)
3. Extracted claims (`user_id`) forwarded as `X-User-Id` header to backend
4. Services do NOT validate JWT themselves — trust the gateway

Protected endpoint: `POST /v1/orders` (via KrakenD `jose` plugin)

### Inter-Service Communication
- **Sync (via gateway):** Client → KrakenD → Service
- **Async (RabbitMQ):** Order Service publishes `order.created` → Kitchen Worker consumes
- Exchange: `order_events` (topic), Queue: `kitchen_order_queue`, Routing key: `order.created`

### Correlation ID Tracing
- KrakenD plugin (`gateway/plugin/`) injects `X-Correlation-ID` if missing
- Each service middleware extracts and propagates it through logs and RabbitMQ headers
- Do not remove this pattern when adding new services

### Circuit Breaker
Order Service wraps Kitchen Service calls with Sony GoBreaker. If adding new cross-service dependencies, consider adding a circuit breaker.

## Adding a New Service

1. Create directory `<name>-service/` with `cmd/main.go` and `internal/` following existing patterns
2. Add `go.mod` with Fiber, GORM, Prometheus middleware, Zerolog
3. Copy logger middleware from an existing service (preserves correlation ID pattern)
4. Add to `docker-compose.yaml`
5. Add Prometheus scrape target to `prometheus.yml`
6. Add gateway routes to `krakend.json`
7. Register with Consul if other services need to discover it

## Adding a New KrakenD Endpoint

Edit `krakend.json`. For a protected endpoint add:
```json
"extra_config": {
  "auth/validator": {
    "alg": "RS256",
    "jwk_local_path": "/etc/krakend/public_key.json",
    "disable_jwk_security": true
  }
}
```

## Logging Convention

Use Zerolog. Every log entry must include `correlation_id`. Pattern from existing middleware:
```go
correlationID := c.Locals("correlationID").(string)
log.Info().
    Str("correlation_id", correlationID).
    Str("service", "my-service").
    Msg("...")
```

## Swagger (Order Service Only)

```bash
cd order-service
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -d .
docker-compose up -d --build api
# View at: http://localhost:3000/swagger/index.html
```

## JWT Key Rotation

If rotating RSA keys:
1. Generate new key pair
2. Replace `private_key.pem` (used by auth-service container at `/app/private_key.pem`)
3. Replace `public_key.pem` and `public_key.json` (JWKS, used by KrakenD at `/etc/krakend/public_key.json`)
4. Rebuild auth-service and gateway containers

## Environment Variables Reference

### Auth Service
```
PORT=3002
DB_URL=postgres://admin:admin@db:5432/auth_db?sslmode=disable
PRIVATE_KEY_PATH=/app/private_key.pem
```

### Order Service
```
PORT=3000
DB_URL=postgres://admin:admin@db:5432/order_db?sslmode=disable
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
CONSUL_ADDRESS=consul:8500
```

### Kitchen Service
```
PORT=3001
DB_URL=postgres://admin:admin@db:5432/kitchen_db?sslmode=disable
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
CONSUL_ADDRESS=consul:8500
SERVICE_ADDRESS=kitchen-api
```

## Testing

No automated test suite exists yet. Manual testing flow:

```bash
# 1. Register
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"pass","email":"t@t.com"}'

# 2. Login
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"pass"}' | jq -r .token)

# 3. Create order
curl -X POST http://localhost:8080/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"customer_id":"c1","total_amount":50.00}'

# 4. Check kitchen ticket
curl http://localhost:8080/v1/kitchen/status/1
```

## Observability URLs

- Grafana: http://localhost:3002 (admin/admin)
- Prometheus: http://localhost:9090
- RabbitMQ: http://localhost:15672 (guest/guest)
- Consul: http://localhost:8500
