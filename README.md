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
  ├──► Kitchen Service (port 3001)                    RabbitMQ (port 5672)
  │      ├── PostgreSQL (kitchen_db)                      │
  │      ├── Consul Registration                          │
  │      └── Kitchen Worker ◄────────────────────────────┘
  │            └── Consumes order.created events
  │
  ├──► Catalog Service (port 3003)
  │      ├── PostgreSQL (catalog_db)
  │      └── Redis (caching)
  │
  └──► Inventory Service (port 3004)
         ├── PostgreSQL (inventory_db)
         ├── HTTP client → Catalog Service (BOM lookup)
         └── Inventory Worker ◄── kitchen.ticket_created ◄── Kitchen Worker

Observability:
  Prometheus (9090) ◄── scrapes all services
  Grafana (3002) ◄── Prometheus + Loki dashboards
  Loki (3100) ◄── Promtail ◄── Docker logs
  Consul UI (8500) ◄── service registry
  RabbitMQ UI (15672) ◄── message broker management
  Redis Insight (8085) ◄── Redis browser
  pgAdmin (5551) ◄── PostgreSQL browser
```

## Services

| Service | Port | Responsibility |
|---------|------|----------------|
| API Gateway (KrakenD) | 8080 | Routing, JWT validation, rate limiting |
| Auth Service | 3002 | User registration, login, JWT issuance, refresh tokens, profile, password management, password reset, group-based roles, permissions, dynamic nav menu |
| Order Service | 3000 | Order CRUD, publishes order events |
| Kitchen Service | 3001 | Kitchen ticket management, consumes order events |
| Kitchen Worker | — | RabbitMQ consumer (separate process) |
| Catalog Service | 3003 | Master data: menus, BOM, add-ons, portions, stations |
| Inventory Service | 3004 | Raw material stock, auto-deduction, low-stock alerts |
| Inventory Worker | — | RabbitMQ consumer for kitchen.ticket_created events |

## Tech Stack

- **Language:** Go 1.25.7
- **Web Framework:** Go Fiber v2
- **ORM:** GORM with PostgreSQL 15
- **Cache:** Redis 7
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
| Redis Insight | http://localhost:8085 | — |
| pgAdmin | http://localhost:5551 | admin@admin.com / admin |
| Order Swagger | http://localhost:3000/swagger/index.html | — |
| Auth Swagger | http://localhost:3002/swagger/index.html | — |

## API Reference

All requests go through the API Gateway at `http://localhost:8080`.

### Authentication

#### Register
```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "password": "secret123", "email": "john@example.com"}'
# Response: {"message":"registered successfully","user":{"id":1,"username":"john","email":"john@example.com","group":"user","roles":["user"]}}
```

#### Login (returns access_token + refresh_token)
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "john", "password": "secret123"}' | jq -r .access_token)
```

#### Refresh Token
```bash
curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token_from_login>"}'
```

#### Get Profile (JWT required)
```bash
curl http://localhost:8080/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

#### Change Password (JWT required)
```bash
curl -X PUT http://localhost:8080/v1/auth/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"current_password": "secret123", "new_password": "newsecret123"}'
```

#### Logout (revoke one refresh token)
```bash
curl -X POST http://localhost:8080/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token>"}'
```

#### Logout All Devices (JWT required)
```bash
curl -X POST http://localhost:8080/v1/auth/logout-all \
  -H "Authorization: Bearer $TOKEN"
```

#### Forgot Password (request reset token)
```bash
curl -X POST http://localhost:8080/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com"}'
# Dev mode: returns reset_token in response body
# Production: integrate email service to send token via email
```

#### Reset Password (using reset token)
```bash
curl -X POST http://localhost:8080/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token": "<reset_token>", "new_password": "newsecret123"}'
# Revokes all refresh tokens to force re-login on all devices
```

#### Get Navigation Menu Config (JWT required)
```bash
curl http://localhost:8080/v1/auth/menu-config \
  -H "Authorization: Bearer $TOKEN"
# Returns menu groups/items filtered by user's permissions (derived from roles)
# Response: {"menu":[{"label":"Auth","permissions":["auth.groups.view",...],"items":[...]},...]}
```

### Orders (requires JWT)

#### Create Order
```bash
curl -X POST http://localhost:8080/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"customer_id": "cust_001", "total_amount": 45.50}'
```

### Kitchen

#### Check Kitchen Status
```bash
curl http://localhost:8080/v1/kitchen/status/{orderId}
```

### Catalog

#### Menu Management
```bash
# List all menus
curl http://localhost:8080/v1/catalog/menus

# Get one menu (includes station info)
curl http://localhost:8080/v1/catalog/menus/1

# Create menu (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"ข้าวผัดกระเพราหมูสับ","price":65.00,"category":"อาหารจานเดียว"}'

# Update / Delete (JWT required)
curl -X PUT  http://localhost:8080/v1/catalog/menus/1 -H "Authorization: Bearer $TOKEN" ...
curl -X DELETE http://localhost:8080/v1/catalog/menus/1 -H "Authorization: Bearer $TOKEN"
```

#### BOM — Multi-level Recipe
```bash
# Add raw ingredient to recipe (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus/1/bom \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ingredient_id":1,"quantity":100}'

# Add another MenuItem as a sub-recipe (JWT required)
# e.g. ปลากระพงทอดน้ำปลา uses ปลากระพงทอด (menu_id=3) as a component
curl -X POST http://localhost:8080/v1/catalog/menus/5/bom \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"sub_menu_item_id":3,"quantity":1}'

# Get structured BOM (nested sub-recipes shown)
curl http://localhost:8080/v1/catalog/menus/1/bom

# Get flat BOM — all sub-recipes expanded, raw ingredients only
# Used by Inventory Service for stock deduction
curl http://localhost:8080/v1/catalog/menus/1/bom/flat
```

#### Choice Groups — Customer Substitution (Case 1)
```bash
# Create choice group e.g. "เลือกเส้น" (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus/1/choices \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"เลือกเส้น","is_required":true,"min_choices":1,"max_choices":1}'

# Add option to group (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus/1/choices/1/options \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ingredient_id":5,"quantity":150,"extra_price":0}'

# List choices
curl http://localhost:8080/v1/catalog/menus/1/choices
```

#### Add-ons — Optional Extras (Case 2)
```bash
# Add optional extra e.g. ไข่ดาว +10฿ (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus/1/addons \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"ingredient_id":3,"quantity":1,"extra_price":10}'

# List add-ons
curl http://localhost:8080/v1/catalog/menus/1/addons
```

#### Portion Sizes — Size Variants (Case 3)
```bash
# Create size variants (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus/1/portions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"ธรรมดา","quantity_multiplier":1.0,"extra_price":0,"is_default":true}'

curl -X POST http://localhost:8080/v1/catalog/menus/1/portions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"พิเศษ","quantity_multiplier":1.5,"extra_price":15}'

# List portions
curl http://localhost:8080/v1/catalog/menus/1/portions
```

#### Kitchen Stations
```bash
# Create station (JWT required)
curl -X POST http://localhost:8080/v1/catalog/stations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"ครัวร้อน","description":"ผัด ทอด"}'

# Assign menu to station (JWT required)
curl -X POST http://localhost:8080/v1/catalog/menus/1/station \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"kitchen_station_id":1}'
```

## Catalog Domain Model

```
MenuItem
  ├── BOMItem[]            recipe entry — one of:
  │     ├── ingredient_id     → raw Ingredient (leaf)
  │     └── sub_menu_item_id  → another MenuItem (sub-recipe, expanded recursively)
  ├── BOMChoiceGroup[]     customer picks one from a group (e.g. เลือกเส้น)
  │     └── BOMChoiceOption[]  each option = ingredient + qty + extra_price
  ├── MenuAddOn[]          optional extras + extra_price (e.g. ไข่ดาว +10฿)
  ├── MenuPortionSize[]    size variants + quantity_multiplier (e.g. พิเศษ ×1.5 +15฿)
  └── KitchenStation[]     which kitchen station handles this menu
```

### Multi-level BOM

BOM items can reference another `MenuItem` as a sub-recipe. Recipes expand recursively — quantities multiply through each level. A menu item sold standalone can simultaneously be a BOM component of another menu.

**Example — เบอร์เกอร์:**
```
เบอร์เกอร์
  ├── sub_menu_item_id: เนื้อวัวทอดกระเทียม (qty: 1) ← sub-recipe, expands to:
  │     ├── เนื้อวัว    50 g
  │     ├── กระเทียม   5 g
  │     └── น้ำมัน     1 tbsp
  ├── ingredient_id: ชีส       (qty: 2 แผ่น)
  └── ingredient_id: มายองเนส  (qty: 2 tbsp)
```

**Example — ปลากระพงทอดน้ำปลา** (reuses an existing MenuItem):
```
ปลากระพงทอดน้ำปลา
  ├── sub_menu_item_id: ปลากระพงทอด (qty: 1) ← also sold as its own menu item
  └── ingredient_id: น้ำปลา (qty: 30 ml)
```

**BOM endpoints:**
| Endpoint | Description |
|----------|-------------|
| `GET /menus/:id/bom` | Structured BOM — shows nested sub-recipe objects |
| `GET /menus/:id/bom/flat` | Flat BOM — all sub-recipes expanded, raw ingredients only (used for stock deduction) |

### Inventory Management
```bash
# Create raw material (links to catalog ingredient)
curl -X POST http://localhost:8080/v1/inventory/materials \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"catalog_ingredient_id":1,"name":"หมูสับ","unit":"g","reorder_point":500}'

# List all materials
curl http://localhost:8080/v1/inventory/materials

# Check low-stock items
curl http://localhost:8080/v1/inventory/materials/low-stock

# Restock (JWT required)
curl -X POST http://localhost:8080/v1/inventory/stock/restock \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"material_id":1,"quantity":2000,"note":"สั่งจากซัพพลายเออร์ A"}'

# Manual deduct by BOM — calls Catalog to get recipe then deducts (JWT required)
curl -X POST http://localhost:8080/v1/inventory/stock/deduct \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"order_id":42,"items":[{"menu_item_id":1,"quantity":2,"portion_multiplier":1.5}]}'

# Transaction history (JWT required)
curl http://localhost:8080/v1/inventory/transactions
curl http://localhost:8080/v1/inventory/transactions/1
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

```
Order Service ──order.created──► Kitchen Worker ──kitchen.ticket_created──► Inventory Worker
                [order_events]                      [kitchen_events]
```

### RabbitMQ Configuration

| Exchange | Routing Key | Producer | Consumer | Queue |
|----------|-------------|----------|----------|-------|
| `order_events` | `order.created` | Order Service | Kitchen Worker | `kitchen_order_queue` |
| `kitchen_events` | `kitchen.ticket_created` | Kitchen Worker | Inventory Worker | `inventory_kitchen_queue` |

### Event Payloads

**order.created**
```json
{ "order_id": 1, "items": "[]" }
```

**kitchen.ticket_created**
```json
{ "order_id": 1, "ticket_id": 5, "items": "[]" }
```

> **Note:** `items` is currently `"[]"` (Order Service TODO). When populated with `[{"menu_item_id":1,"quantity":2,"portion_multiplier":1.0}]`, Inventory Worker will auto-deduct stock via BOM lookup from Catalog Service.

Correlation ID is passed via RabbitMQ message `CorrelationId` header for end-to-end request tracing.

## Authentication & Security

- **Password hashing:** bcrypt (cost 10)
- **JWT algorithm:** RS256 (RSA 2048-bit)
- **Access token expiry:** 24 hours
- **Refresh token expiry:** 7 days (stored as SHA-256 hash in DB)
- **Password reset token expiry:** 15 minutes (stored as SHA-256 hash in DB)
- **Token validation:** Performed at gateway level before requests reach backend services
- **Refresh flow:** Client sends refresh_token → Auth Service rotates (delete old, issue new pair)
- **Password reset flow:** Request token via email → validate token → set new password → revoke all refresh tokens
- **Key rotation:** Replace `private_key.pem`, `public_key.pem`, and `public_key.json` (JWKS format)
- **Authorization model:** Group-based roles with permissions — each user belongs to a Group, each Group has multiple Roles (many-to-many), and each Role has multiple Permissions (many-to-many via `role_permissions`). Permissions use `module.resource.action` format (e.g. `auth.groups.view`, `catalog.menus.manage`). New registrations default to the `user` group. JWT claims include `group` (string) and `roles` (string array). Permissions are resolved server-side from roles.
- **Dynamic navigation menu:** Menu config stored in DB (`nav_groups`/`nav_items` with FK to `permissions`). Frontend fetches `GET /v1/auth/menu-config` — backend filters by user's permissions derived from roles.

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

### Catalog Service
```env
PORT=3003
DB_URL=postgres://admin:admin@db:5432/catalog_db?sslmode=disable
REDIS_URL=redis:6379
```

### Inventory Service
```env
PORT=3004
DB_URL=postgres://admin:admin@db:5432/inventory_db?sslmode=disable
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
CATALOG_SERVICE_URL=http://catalog-service-api:3003
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
- Catalog Service: `catalog-service-api:3003/metrics`
- Inventory Service: `inventory-service-api:3004/metrics`
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

## Frontend (Next.js)

The frontend is a Next.js application located in `front-end/`. It provides:

| Page | Path | Description |
|------|------|-------------|
| Login | `/auth/login` | Username/password login via NextAuth.js |
| Register | `/auth/register` | New user registration |
| Forgot Password | `/auth/forgot-password` | Request password reset token by email |
| Reset Password | `/auth/reset-password` | Set new password using reset token (supports `?token=` URL param) |
| Dashboard | `/dashboard` | Protected page (requires authentication) |

The navigation menu is **dynamic and permission-based** — the navbar fetches `GET /api/auth/menu-config` on login, which returns only the menu items the user is allowed to see based on their role permissions.

API proxy routes (`app/api/auth/`) forward requests to the backend Auth Service.

> **Dev mode note:** The forgot-password endpoint returns the `reset_token` in the response body. Remove this field when integrating a real email service.

## Project Structure

```
food-delivery/
├── auth-service/
│   ├── cmd/main.go
│   └── internal/
│       ├── handler/        # HTTP handlers
│       ├── middleware/      # Logger
│       ├── model/           # User, Group, Role, Permission, NavGroup, NavItem, RefreshToken, PasswordResetToken
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
├── catalog-service/
│   ├── cmd/main.go
│   └── internal/
│       ├── handler/         # menu, ingredient, bom, choice, addon, portion, station
│       ├── middleware/      # Logger + Correlation ID
│       ├── model/           # MenuItem, Ingredient, BOMItem, BOMChoiceGroup,
│       │                    #   BOMChoiceOption, MenuAddOn, MenuPortionSize,
│       │                    #   KitchenStation, MenuStationMapping
│       ├── repository/      # Database layer (7 repositories)
│       └── service/         # Business logic (7 services)
├── front-end/
│   ├── app/
│   │   ├── api/auth/        # API proxy routes (NextAuth, forgot/reset password)
│   │   ├── auth/            # Auth pages (login, register, forgot/reset password)
│   │   └── dashboard/       # Protected dashboard page
│   ├── middleware.ts         # NextAuth middleware
│   └── package.json
├── gateway/
│   └── plugin/              # Correlation ID injector plugin
├── krakend.json             # API Gateway routing & JWT config
├── docker-compose.yaml      # Full stack orchestration
├── prometheus.yml           # Metrics scrape config
├── promtail-config.yaml     # Log shipping config
├── inventory-service/
│   ├── cmd/
│   │   ├── main.go          # API server (port 3004)
│   │   └── worker/main.go   # Consumes kitchen.ticket_created, deducts stock
│   └── internal/
│       ├── catalog/         # HTTP client for Catalog Service BOM lookup
│       ├── handler/         # material, stock, transaction handlers
│       ├── middleware/      # Logger
│       ├── model/           # RawMaterial, StockTransaction
│       ├── repository/      # material_repository, transaction_repository
│       └── service/         # material_service, stock_service (DeductByBOM)
├── init.sql                 # Creates order_db, kitchen_db, auth_db, catalog_db, inventory_db
├── private_key.pem          # RSA private key (JWT signing)
├── public_key.pem           # RSA public key
└── public_key.json          # JWKS format (KrakenD JWT validation)
```

## Database Schema

Tables are auto-migrated by GORM on service startup.

**users** (auth_db)
```
id, username (unique), password (bcrypt), email (unique), group_id (FK→groups), created_at, updated_at, deleted_at
```

**groups** (auth_db)
```
id, name (unique), created_at, updated_at, deleted_at
```

**roles** (auth_db)
```
id, name (unique), created_at, updated_at, deleted_at
```

**permissions** (auth_db)
```
id, name (unique, format: module.resource.action), description, created_at, updated_at, deleted_at
```

**group_roles** (auth_db) — join table (many-to-many)
```
group_id (FK→groups), role_id (FK→roles)
```

**role_permissions** (auth_db) — join table (many-to-many)
```
role_id (FK→roles), permission_id (FK→permissions)
```

**nav_groups** (auth_db) — navigation menu groups
```
id, label (unique), sort_order, created_at, updated_at, deleted_at
```

**nav_items** (auth_db) — navigation menu items
```
id, nav_group_id (FK→nav_groups), label, href, sort_order, created_at, updated_at, deleted_at
```

**nav_group_permissions** (auth_db) — join table
```
nav_group_id (FK→nav_groups), permission_id (FK→permissions)
```

**nav_item_permissions** (auth_db) — join table
```
nav_item_id (FK→nav_items), permission_id (FK→permissions)
```

**Seeded groups:**
| Group | Roles |
|-------|-------|
| user | user |
| customer | customer, user |
| rider | rider, user |
| merchant | merchant, user |
| admin | admin, merchant, rider, customer, user |

**Seeded permissions (18):**
| Module | Permissions |
|--------|------------|
| Auth | `auth.groups.view`, `auth.groups.manage`, `auth.roles.view`, `auth.roles.manage`, `auth.users.view`, `auth.users.manage` |
| Catalog | `catalog.menus.view`, `catalog.menus.manage`, `catalog.ingredients.view`, `catalog.ingredients.manage` |
| Kitchen | `kitchen.view`, `kitchen.manage` |
| Order | `orders.view`, `orders.manage`, `orders.create`, `orders.queue.view` |
| Inventory | `inventory.view`, `inventory.manage` |

**Role → Permission mapping:**
| Role | Permissions |
|------|------------|
| admin | all 18 |
| merchant | catalog.*, kitchen.*, orders.view/manage/queue, inventory.* |
| rider | orders.queue.view, orders.view |
| customer | orders.create, orders.view |
| user | (none) |

**refresh_tokens** (auth_db)
```
id, user_id (FK→users), token_hash (unique, SHA-256), expires_at, created_at
```

**password_reset_tokens** (auth_db)
```
id, user_id (FK→users), token_hash (unique, SHA-256), expires_at (15 min), created_at
```

**orders** (order_db)
```
id, customer_id, total_amount (decimal 10,2), status (default: Pending), created_at, updated_at
```

**kitchen_tickets** (kitchen_db)
```
id, order_id (unique), items (JSON), status (Received|Cooking|Ready), created_at
```

**menu_items** (catalog_db)
```
id, name (unique), description, price (decimal 10,2), category, is_available, created_at, updated_at
```

**ingredients** (catalog_db)
```
id, name (unique), unit (g/ml/piece/...), created_at, updated_at
```

**bom_items** (catalog_db)
```
id, menu_item_id, ingredient_id (nullable), sub_menu_item_id (nullable), quantity (decimal 10,3), created_at, updated_at
-- exactly one of ingredient_id or sub_menu_item_id must be non-null
```

**bom_choice_groups** (catalog_db)
```
id, menu_item_id, name, is_required, min_choices, max_choices, created_at, updated_at
```

**bom_choice_options** (catalog_db)
```
id, group_id, ingredient_id, quantity, extra_price (decimal 10,2), created_at, updated_at
```

**menu_add_ons** (catalog_db)
```
id, menu_item_id, ingredient_id, quantity, extra_price, is_available, created_at, updated_at
```

**menu_portion_sizes** (catalog_db)
```
id, menu_item_id, name, quantity_multiplier (decimal 5,2), extra_price, is_default, created_at, updated_at
```

**kitchen_stations** (catalog_db)
```
id, name (unique), description, created_at, updated_at
```

**menu_station_mappings** (catalog_db)
```
menu_item_id (PK), kitchen_station_id (PK)
```

**raw_materials** (inventory_db)
```
id, catalog_ingredient_id (nullable, links to catalog), name (unique), unit, current_stock (decimal 12,3), reorder_point, created_at, updated_at
```

**stock_transactions** (inventory_db)
```
id, raw_material_id, quantity_change (decimal 12,3), type (RESTOCK|DEDUCTION|ADJUSTMENT), order_id (nullable), correlation_id, note, created_at
```

## Development

### Regenerate Swagger Docs
```bash
# Order Service
cd order-service
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -d .
docker-compose up -d --build api

# Auth Service
cd auth-service
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -d .
docker-compose up -d --build auth-service
```

### Run Unit Tests
```bash
# Auth Service (39 tests — handler + service layers)
cd auth-service && go test ./internal/... -v
```

### Rebuild a Single Service
```bash
docker-compose up -d --build auth-service
docker-compose up -d --build api           # order-service
docker-compose up -d --build kitchen-api
docker-compose up -d --build catalog-service
docker-compose up -d --build inventory-api
docker-compose up -d --build inventory-worker
```
