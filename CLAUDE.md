# CLAUDE.md — Food Delivery Microservices

Development guide for Claude Code when working in this repository.

## Project Overview

Microservices-based food delivery backend written in Go. Five main services (auth, order, kitchen, catalog, inventory) communicate through RabbitMQ events and are exposed via KrakenD API gateway. Full observability via Prometheus + Grafana + Loki.

## Repository Structure

```
food-delivery/
├── auth-service/        # JWT issuance, user management, password reset
├── front-end/           # Next.js frontend (login, register, forgot/reset password, dashboard)
├── order-service/       # Order CRUD + RabbitMQ publisher
├── kitchen-service/     # Kitchen tickets + RabbitMQ consumer worker
├── catalog-service/     # Master data: menus, BOM, add-ons, portions, stations
├── inventory-service/   # Stock tracking, auto-deduction, low-stock alerts
├── gateway/             # KrakenD plugin (correlation ID injector)
├── docker-compose.yaml  # Full stack (19 containers)
├── krakend.json         # Gateway routing + JWT validation config
├── prometheus.yml       # Metrics scrape config
├── promtail-config.yaml # Loki log shipping
├── init.sql             # Creates 5 PostgreSQL databases
├── private_key.pem      # RSA-2048 private key (JWT signing)
├── public_key.pem       # RSA public key
└── public_key.json      # JWKS format for KrakenD
```

Each service is a standalone Go module with its own `go.mod`.

## Tech Stack

- **Go 1.25.7**, **Go Fiber v2**, **GORM**
- **PostgreSQL 15** — one DB per service (`order_db`, `kitchen_db`, `auth_db`, `catalog_db`, `inventory_db`)
- **Redis 7** — caching layer (currently used by Catalog Service)
- **RabbitMQ 3** — async event bus between Order and Kitchen
- **KrakenD** — API gateway with JWT RS256 validation
- **Consul** — service discovery (Kitchen registers; Order can discover)
- **Prometheus + Grafana + Loki + Promtail** — observability
- **Jaeger** — distributed tracing (OpenTelemetry / OTLP gRPC)
- **Zerolog** — structured JSON logging
- **Sony GoBreaker** — circuit breaker in Order Service
- **golang-jwt/jwt/v5** — JWT generation (Auth Service)

## Running the Stack

```bash
# Start everything
docker-compose up -d --build

# Rebuild one service
docker-compose up -d --build api              # order-service
docker-compose up -d --build auth-service
docker-compose up -d --build kitchen-api
docker-compose up -d --build kitchen-worker
docker-compose up -d --build catalog-service
docker-compose up -d --build inventory-api
docker-compose up -d --build inventory-worker
```

## Service Ports

| Service | Internal Port | Exposed Port |
|---------|--------------|--------------|
| API Gateway | 8080 | 8080 |
| Order Service | 3000 | 3000 |
| Kitchen Service | 3001 | 3001 |
| Auth Service | 3002 | 3005 |
| Catalog Service | 3003 | 3003 |
| Inventory Service | 3004 | 3004 |
| PostgreSQL | 5432 | 5555 |
| pgAdmin | 80 | 5551 |
| RabbitMQ | 5672 | 5672 |
| RabbitMQ UI | 15672 | 15672 |
| Redis | 6379 | 6379 |
| Redis Insight | 5540 | 8085 |
| Consul | 8500 | 8500 |
| Prometheus | 9090 | 9090 |
| Grafana | 3000 | 3002 |
| Loki | 3100 | 3100 |
| Jaeger UI | 16686 | 16686 |
| Jaeger OTLP gRPC | 4317 | 4317 |

## Key Architectural Decisions

### Each Service Has Its Own Database
Never share a database between services. Each service owns its schema and connects to its own DB (`order_db`, `kitchen_db`, `auth_db`, `catalog_db`, `inventory_db`).

### Authentication Flow
1. Auth Service issues RS256 JWT (access token + refresh token) signed with `private_key.pem`
2. KrakenD validates all protected requests using `public_key.json` (JWKS)
3. Extracted claims (`user_id`) forwarded as `X-User-Id` header to backend
4. Services do NOT validate JWT themselves — trust the gateway

**Auth endpoints:**
| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/v1/auth/register` | POST | No | Register new user |
| `/v1/auth/login` | POST | No | Login, returns access_token + refresh_token |
| `/v1/auth/refresh` | POST | No | Refresh access token using refresh_token |
| `/v1/auth/logout` | POST | No | Revoke a specific refresh token |
| `/v1/auth/logout-all` | POST | JWT | Revoke all refresh tokens for user |
| `/v1/auth/profile` | GET | JWT | Get current user profile |
| `/v1/auth/password` | PUT | JWT | Change password |
| `/v1/auth/forgot-password` | POST | No | Request password reset token (returns token in dev mode) |
| `/v1/auth/reset-password` | POST | No | Reset password using token (revokes all refresh tokens) |

Protected endpoints: `POST /v1/orders`, write endpoints under `/v1/catalog/*`, `/v1/auth/profile`, `/v1/auth/logout-all`, `/v1/auth/password`

### Inter-Service Communication
- **Sync (via gateway):** Client → KrakenD → Service
- **Async (RabbitMQ):**
  - Order Service publishes `order.created` → Kitchen Worker consumes → creates KitchenTicket
  - Kitchen Worker publishes `kitchen.ticket_created` → Inventory Worker consumes → deducts stock via BOM
- Exchanges: `order_events` (topic), `kitchen_events` (topic)
- Queues: `kitchen_order_queue`, `inventory_kitchen_queue`

### Inventory Service — Stock Tracking
Inventory Service owns stock levels for raw materials. It does not duplicate ingredient master data — it links via `catalog_ingredient_id` to ingredients in Catalog Service.

**Auto-deduction flow:**
1. Kitchen Worker creates KitchenTicket and publishes `kitchen.ticket_created` to `kitchen_events` exchange
2. Inventory Worker consumes the event and calls Catalog Service HTTP API (`GET /api/v1/catalog/menus/{id}/bom/flat`) to resolve BOM
3. The `/bom/flat` endpoint recursively expands all sub-recipes and returns only raw ingredients with multiplied quantities
4. For each ingredient, finds the matching `RawMaterial` by `catalog_ingredient_id` and deducts stock
5. If `current_stock < reorder_point`, logs `Warn` alert with `alert: LOW_STOCK` field

**Note:** Auto-deduction requires Order Service to include `menu_item_ids` in the RabbitMQ event payload (currently a TODO). Manual deduction is available via `POST /v1/inventory/stock/deduct`.

### Catalog Service — Master Data
Catalog Service is the source of truth for menu data. It does NOT receive events — other services query it via the gateway if needed.

**Domain model:**
```
MenuItem
  ├── BOMItem[]           — recipe entries (ingredient OR sub-recipe)
  │     ├── ingredient_id    → raw Ingredient (leaf node)
  │     └── sub_menu_item_id → another MenuItem whose BOM is expanded recursively
  ├── BOMChoiceGroup[]    — customer-selectable ingredient groups (e.g. เลือกเส้น)
  │     └── BOMChoiceOption[]
  ├── MenuAddOn[]         — optional extras with extra_price (e.g. ไข่ดาว +10฿)
  ├── MenuPortionSize[]   — size variants with quantity_multiplier (e.g. พิเศษ ×1.5)
  └── KitchenStation[]    — which kitchen station handles this menu item
```

**Multi-level BOM:** A `BOMItem` can reference either a raw `Ingredient` (leaf) or another `MenuItem` as a sub-recipe. Sub-recipes expand recursively — quantities multiply through each level. The same MenuItem can be sold standalone AND used as a sub-recipe in another menu (e.g. ปลากระพงทอด → ปลากระพงทอดน้ำปลา).

**BOM endpoints:**
- `GET /api/v1/catalog/menus/:id/bom` — structured BOM (shows nested `sub_menu_item` objects)
- `GET /api/v1/catalog/menus/:id/bom/flat` — fully-expanded flat list of raw ingredients (used by Inventory Service for stock deduction)

**Constraint:** Each `BOMItem` must have exactly one of `ingredient_id` or `sub_menu_item_id`. Setting both or neither returns `400 Bad Request`. Circular references are detected at insert time and return `422 Unprocessable Entity`.

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
7. Add database to `init.sql`
8. Register with Consul if other services need to discover it

## Adding a New KrakenD Endpoint

Edit `krakend.json`. For a protected endpoint add:
```json
{
  "endpoint": "/v1/your-endpoint",
  "method": "GET",
  "output_encoding": "no-op",
  "input_headers": ["Authorization", "Content-Type", "X-User-Id"],
  "extra_config": {
    "auth/validator": {
      "alg": "RS256",
      "jwk_local_path": "/etc/krakend/public_key.json",
      "disable_jwk_security": true,
      "propagate_claims": [
        ["user_id", "X-User-Id"]
      ]
    }
  },
  "backend": [{
    "url_pattern": "/api/v1/your-endpoint",
    "host": ["http://your-service:port"],
    "encoding": "no-op"
  }]
}
```

**Important (KrakenD 2.13):** Propagated claim headers (e.g. `X-User-Id`) MUST be listed in `input_headers` at the endpoint level, otherwise they will not be forwarded to the backend. Always include `"X-User-Id"` in `input_headers` when using `propagate_claims`.

## Logging Convention

Use Zerolog. Every log entry must include `correlation_id`. Pattern from existing middleware:
```go
correlationID := c.Locals("correlationID").(string)
log.Info().
    Str("correlation_id", correlationID).
    Str("service", "my-service").
    Msg("...")
```

## Swagger

```bash
# Order Service
cd order-service
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -d .
docker-compose up -d --build api
# View at: http://localhost:3000/swagger/index.html

# Auth Service
cd auth-service
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -d .
docker-compose up -d --build auth-service
# View at: http://localhost:3002/swagger/index.html
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

### Catalog Service
```
PORT=3003
DB_URL=postgres://admin:admin@db:5432/catalog_db?sslmode=disable
REDIS_URL=redis:6379
```

### Inventory Service
```
PORT=3004
DB_URL=postgres://admin:admin@db:5432/inventory_db?sslmode=disable
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
CATALOG_SERVICE_URL=http://catalog-service-api:3003
```

## Testing

### Unit Tests

Auth Service has unit tests for handler and service layers:

```bash
cd auth-service
go test ./internal/... -v
```

Test coverage: Register, Login, Refresh, Logout, LogoutAll, GetProfile, ChangePassword — both success and error cases (39 tests).

### Manual Testing Flow

```bash
# 1. Register
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"testpass1","email":"t@t.com"}'

# 2. Login (returns access_token + refresh_token)
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"testpass1"}' | jq -r .access_token)

# 3. Get profile (JWT required)
curl http://localhost:8080/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"

# 4. Refresh token
REFRESH=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"testpass1"}' | jq -r .refresh_token)
curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}"

# 5. Change password (JWT required)
curl -X PUT http://localhost:8080/v1/auth/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"current_password":"testpass1","new_password":"newpass123"}'

# 6. Logout (revoke one refresh token)
curl -X POST http://localhost:8080/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}"

# 7. Logout all devices (JWT required)
curl -X POST http://localhost:8080/v1/auth/logout-all \
  -H "Authorization: Bearer $TOKEN"

# 8. Forgot password (returns reset_token in dev mode)
curl -X POST http://localhost:8080/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"t@t.com"}'

# 9. Reset password (using token from step 8)
curl -X POST http://localhost:8080/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token":"<reset_token>","new_password":"resetpass123"}'

# 10. Create order (JWT required)
curl -X POST http://localhost:8080/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"customer_id":"c1","total_amount":50.00}'

# 11. Check kitchen ticket
curl http://localhost:8080/v1/kitchen/status/1

# 12. Create a catalog menu item (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"ข้าวผัดกระเพราหมูสับ","price":65.00,"category":"อาหารจานเดียว"}'
```

## Observability URLs

- Grafana: http://localhost:3002 (admin/admin)
- Prometheus: http://localhost:9090
- RabbitMQ: http://localhost:15672 (guest/guest)
- Consul: http://localhost:8500
- Redis Insight: http://localhost:8085
- Jaeger: http://localhost:16686
- pgAdmin: http://localhost:5551 (admin@admin.com/admin)
