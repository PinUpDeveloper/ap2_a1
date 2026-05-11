# Architecture Diagram — Assignment 4

```text
Client / Postman
    |
    v
Order Service (:8080)
    |-- Redis Rate Limiter (bonus)
    |-- GET /orders/:id
    |      |-- Redis GET order:<id>
    |      |-- cache HIT  -> return cached order
    |      |-- cache MISS -> PostgreSQL -> Redis SET with TTL -> return order
    |
    |-- POST /orders
           |-- PostgreSQL INSERT order
           |-- gRPC authorize payment
           |-- PostgreSQL UPDATE status + updated_at
           |-- Redis cache invalidation/update
           v
Payment Service (:8081 / gRPC :50051)
    |-- PostgreSQL payments
    |-- RabbitMQ publish payment.completed
           v
RabbitMQ queue: payment.completed
           v
Notification Service Worker
    |-- Redis GET notification:processed:<event_id> for idempotency
    |-- EmailSender interface
    |      |-- Simulated provider by default
    |      |-- SMTP provider optional
    |-- Retry with exponential backoff: 2s, 4s, 8s, 16s, 32s
    |-- Redis SET notification:processed:<event_id> = sent/failed
```
