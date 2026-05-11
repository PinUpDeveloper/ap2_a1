# AP2 – Assignment 4: Performance Optimization & External Integrations

Built on top of Assignment 3 (Order, Payment, Notification microservices).

## What was added

### 1. Redis Caching – Order Service (Cache-Aside Pattern)

**Interface:** `domain.OrderCache` keeps the use-case layer unaware of Redis.  
**Adapter:** `repository/redis.OrderCache` wraps `go-redis`.

**Read path (`GET /orders/:id`):**
1. Check Redis – if hit, return immediately (logged as `[cache] HIT`).
2. On miss, query Postgres, then write result to Redis with a 5-minute TTL.

**Invalidation strategy (atomic):**  
Immediately after any `UpdateStatus` DB call (order paid, failed, or cancelled), the corresponding Redis key is deleted via `cache.Delete`. This prevents stale data from being served even within the TTL window.

```
GET /orders/:id
  └─ Redis HIT  → return cached order
  └─ Redis MISS → DB query → set Redis (TTL 5min) → return
```

### 2. External Provider Adapter – Notification Service

`provider.EmailSender` is the interface all adapters implement.

| Adapter | Description |
|---|---|
| `SimulatedEmailSender` | Logs the send, sleeps 200–800 ms (latency), fails ~30% of the time (tests retry logic) |
| `SMTPEmailSender` | Real SMTP via `net/smtp` |

Switch via env var: `PROVIDER_MODE=SIMULATED` (default) or `PROVIDER_MODE=REAL`.

### 3. Background Job Worker – Notification Service

`worker.NotificationWorker` is called from the RabbitMQ consumer **in a goroutine** so the consumer loop is never blocked by a slow provider.

**Idempotency:**  
Before sending, the worker checks Redis for key `notification:processed:<event_id>`.  
On success it writes `sent` (TTL 24h); on final failure it writes `failed`.  
Duplicate deliveries are silently skipped.

**Retry with Exponential Backoff:**
```
attempt 1 → fail → wait 2s
attempt 2 → fail → wait 4s
attempt 3 → fail → wait 8s
attempt 4 → fail → wait 16s
attempt 5 → fail → give up (mark "failed" in Redis)
```

### 4. BONUS – Redis Rate Limiter Middleware

`middleware.RateLimiter` uses Redis `INCR` + `EXPIRE` to count requests per client IP.  
Default: **10 requests / 60 seconds**.  
Returns `HTTP 429 Too Many Requests` with `X-RateLimit-*` headers when exceeded.

## Architecture

```
Client
  │
  ▼
Order Service (HTTP :8080)
  ├── middleware: RateLimiter (Redis INCR)          ← BONUS
  ├── usecase: OrderUsecase
  │     ├── Cache-aside read  (Redis GET)
  │     ├── DB read/write     (Postgres)
  │     └── Cache invalidate  (Redis DEL on update)
  └── gRPC client → Payment Service (:50051)

Payment Service
  └── Postgres
  └── RabbitMQ publisher → queue: payment.completed

Notification Service (worker)
  ├── RabbitMQ consumer
  └── NotificationWorker (goroutine per message)
        ├── Idempotency check  (Redis GET notification:processed:<id>)
        ├── EmailSender.Send() with Exponential Backoff (up to 5 retries)
        └── Record result      (Redis SET notification:processed:<id>)

Infrastructure
  ├── Redis   :6379  (shared by Order cache, Rate limiter, Notification idempotency)
  ├── RabbitMQ :5672
  ├── order-db  Postgres :5433
  └── payment-db Postgres :5434
```

## Running

```bash
docker compose up --build
```

### Quick test

```bash
# Create order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"c1","item_name":"Book","amount":1999}'

# Get order (first call: DB; second call: Redis cache)
curl http://localhost:8080/orders/<id>

# Test rate limiter (run 11+ times quickly → 429)
for i in $(seq 1 12); do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/orders/fake; done
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `CACHE_TTL_SECONDS` | `300` | Order cache TTL (5 min) |
| `RATE_LIMIT_REQUESTS` | `10` | Max requests per window |
| `RATE_LIMIT_WINDOW_SECONDS` | `60` | Rate limit window |
| `PROVIDER_MODE` | `SIMULATED` | `SIMULATED` or `REAL` |
| `SMTP_HOST` | `` | SMTP host (REAL mode) |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USER` | `` | SMTP username |
| `SMTP_PASSWORD` | `` | SMTP password |
| `SMTP_FROM` | `noreply@example.com` | Sender address |
