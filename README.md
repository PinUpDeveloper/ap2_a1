# ap2_a1 - Микросервисный backend

Backend-проект на Go для обработки заказов, платежей и уведомлений. Система разделена на несколько сервисов, которые взаимодействуют между собой через REST API, gRPC и RabbitMQ.

## Стек

- Go
- Gin
- gRPC
- Protocol Buffers
- PostgreSQL
- Redis
- RabbitMQ
- Docker / Docker Compose

## Архитектура

Проект состоит из трех основных микросервисов:

- `order-service` - создание, получение и отмена заказов.
- `payment-service` - обработка платежей и публикация событий.
- `notification-service` - асинхронная обработка уведомлений.

Общая схема работы:

```text
Client
  |
  v
Order Service
  |
  | gRPC
  v
Payment Service
  |
  | RabbitMQ
  v
Notification Service
```

## Основной функционал

- REST API для работы с заказами.
- gRPC-коммуникация между сервисами.
- Protocol Buffers для описания контрактов.
- PostgreSQL для хранения заказов и платежей.
- Redis-кэширование заказов по cache-aside pattern.
- Redis rate limiting для ограничения количества запросов.
- RabbitMQ для асинхронной отправки событий.
- Background worker для обработки уведомлений.
- Retry logic с exponential backoff.
- Идемпотентность обработки уведомлений через Redis.
- Docker Compose для запуска всей инфраструктуры.

## Структура проекта

```text
.
├── order-service/          # сервис заказов
├── payment-service/        # сервис платежей
├── notification-service/   # сервис уведомлений
├── contracts/              # сгенерированные gRPC-контракты
├── protos/                 # .proto файлы
├── migrations/             # SQL-миграции для PostgreSQL
├── docker-compose.yml
├── Dockerfile.order
├── Dockerfile.payment
└── Dockerfile.notification
```

## Запуск

Для запуска всех сервисов и инфраструктуры:

```bash
docker compose up --build
```

После запуска будут доступны:

- Order Service: `http://localhost:8080`
- Payment Service: `http://localhost:8081`
- RabbitMQ Management UI: `http://localhost:15672`
- Redis: `localhost:6379`
- Order DB: `localhost:5433`
- Payment DB: `localhost:5434`

## Примеры запросов

Создание заказа:

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"c1","item_name":"Book","amount":1999}'
```

Получение заказа:

```bash
curl http://localhost:8080/orders/<order_id>
```

Получение последних заказов:

```bash
curl http://localhost:8080/orders/recent?limit=5
```

Отмена заказа:

```bash
curl -X PATCH http://localhost:8080/orders/<order_id>/cancel
```

## Что было реализовано

- Разделение backend на слои `domain`, `usecase`, `repository`, `transport`.
- REST handlers на Gin.
- gRPC client/server для связи сервисов.
- Repository layer для работы с PostgreSQL.
- Redis cache layer для заказов.
- RabbitMQ publisher/consumer.
- Асинхронный notification worker.
- Simulated и SMTP email provider.
- Конфигурация сервисов через environment variables.

