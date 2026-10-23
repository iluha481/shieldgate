# ShieldGate — Distributed API Gateway & Rate Limiting SaaS

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)

**ShieldGate** is a high-performance, distributed API Gateway and Reverse Proxy written in Go. It protects upstream microservices from traffic surges and DDoS attacks using atomic Redis Lua scripts, distributed rate limiting, and real-time WebSocket analytics.

---

##  Key Features

- **High-Performance Reverse Proxy:** Built on Go's native `net/http/httputil` utilizing Go 1.26+ `Rewrite` API and connection pooling (`http.Transport` Keep-Alive).
- **Atomic Sliding Window Rate Limiting:** Implemented via **Redis Lua scripts** to guarantee zero race conditions and sub-millisecond execution under high concurrency.
- ** wo-Tier Caching (Cache-Aside):** Fast route resolution and API-key verification cached in Redis with jitter (protects against Cache Avalanche) and immediate cache invalidation on mutations.
- **Real-time WebSocket Analytics:** In-memory metrics collection via `sync/atomic` (zero GC pressure) with live streaming to a web dashboard.
- **Enterprise-Grade Security:** Cryptographically secure API keys (`sg_live_...`) hashed with SHA-256 before storage; passwords hashed using `bcrypt`.
- **Clean Architecture:** Strict separation of concerns (Domain, Service, Repository, Handler) with 100% decoupling via Go interfaces.

---

## Architecture Overview

```text
                  [ Incoming Client Request ]
                               │
                               ▼
                    ┌─────────────────────┐
                    │  ShieldGate Gateway │
                    └──────────┬──────────┘
                               │
            ┌──────────────────┴──────────────────┐
            ▼                                     ▼
   [ Fast Cache Lookup ]                 [ Atomic Rate Limiting ]
   Redis (Routes & Keys)                 Redis Lua (Sliding Window)
            │                                     │
            └──────────────────┬──────────────────┘
                               │
                   (If Allowed: HTTP 200)
                               │
                               ▼
                    [ Target Upstream API ]
```

---

##  Tech Stack

- **Backend:** Go 1.22+ (`net/http`, `sync/atomic`, `gorilla/websocket`)
- **Database:** PostgreSQL 16 (via binary driver `jackc/pgx/v5` with connection pooling)
- **Cache & Limits:** Redis 7 (Sorted Sets `ZSet`, atomic Lua scripting)
- **Auth:** JWT (`golang-jwt/jwt/v5`) & `golang.org/x/crypto/bcrypt`
- **DevOps:** Docker, Docker Compose, Multi-stage builds, GNU Make

---

##  Quick Start (Run with Docker)

### 1. Clone the repository
```bash
git clone https://github.com/your-username/shieldgate.git
cd shieldgate
```

### 2. Configure environment
```bash
cp .env.example .env
```

### 3. Start all services (App + PostgreSQL + Redis + Mock Upstream)
```bash
make up
```
*Or using docker-compose directly:*
```bash
docker-compose up --build -d
```

The gateway will be running at `http://localhost:8080`.

---

##  API Usage & Quick Test

### 1. Register a new user
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@shieldgate.io", "password":"password123"}'
```

### 2. Login to get JWT Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@shieldgate.io", "password":"password123"}'
```
*Export your token:* `TOKEN="your_jwt_token_here"`

### 3. Register an Upstream Service (Limit: 3 req/sec)
```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Echo API",
    "path_prefix": "/echo",
    "target_url": "http://mock-upstream:5678",
    "rate_limit_rps": 3,
    "burst_capacity": 3
  }'
```

### 4. Generate an API Key
```bash
curl -X POST http://localhost:8080/api/v1/services/1/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "Client App"}'
```
*Export your key:* `API_KEY="sg_live_..."`

### 5. Test Rate Limiter (Simulation)
Send 10 requests instantly:
```bash
for i in {1..10}; do 
  curl -s -o /dev/null -w "%{http_code}\n" -H "X-API-Key: $API_KEY" http://localhost:8080/echo/hello
done
```
**Expected Output:**
```text
200
200
200
429
429
429
429
...
```

---

##  Project Structure

```text
.
├── cmd/shieldgate/         # Main application entrypoint
├── internal/
│   ├── app/                # Application bootstrap & graceful shutdown
│   ├── config/             # Environment configuration
│   ├── controller/         # HTTP Handlers (v1) & Middlewares (Auth, Metrics)
│   ├── domain/             # Core business entities & domain errors
│   ├── repository/         # PostgreSQL & Redis cache repositories
│   └── service/            # Core business use cases (Auth, Routes, Gateway)
├── pkg/                    # Reusable packages (JWT, Postgres wrapper)
├── web/                    # Embedded frontend dashboard (HTML, Tailwind, Chart.js)
├── docker-compose.yml      # Local infrastructure definition
├── Dockerfile              # Minimal multi-stage production build (~20MB)
└── Makefile                # Quick task automation
```

