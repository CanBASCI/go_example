# Go Example – Go Fiber with Event-Driven Saga

This project is a microservices example in **Go** with **Go Fiber v3**, matching the same architecture as the Spring Boot 4 reference project. It includes an API Gateway, User Service, Order Service, and Kafka-based event-driven saga (choreography).

## Architecture

- **API Gateway** (port 8080) – Go Fiber reverse proxy with round-robin for user-service
- **User Service** (ports 8081, 8082) – User and balance management, Kafka event consumer
- **Order Service** (port 8091) – Order and saga orchestration, Kafka producer/consumer
- **Event-Driven Saga** – Asynchronous communication and compensation via Apache Kafka
- **PostgreSQL** – Per-service databases

## Tech Stack

- **Go 1.22+** (Fiber v3 officially requires Go 1.25+; project uses go 1.22, upgrade Go if needed)
- **Go Fiber v3** – Web framework (latest major)
- **pgx v5** – PostgreSQL driver
- **segmentio/kafka-go** – Kafka client
- **Docker & Docker Compose**

## Project Layout (Go standard layout)

```
go_example/
├── cmd/
│   ├── gateway/          # API Gateway
│   ├── user-service/     # User service
│   └── order-service/    # Order service
├── internal/
│   └── events/           # Shared Kafka event types
├── go.mod
├── docker-compose.yml
└── README.md
```

## Prerequisites

- Go 1.22+ (Go 1.25 recommended for Fiber v3)
- Docker and Docker Compose

## Run with Docker Compose

```bash
# Resolve dependencies (creates go.sum)
go mod tidy

# Start all services
docker compose up --build
```

Gateway: http://localhost:8080  
User Service 1: http://localhost:8081  
User Service 2: http://localhost:8082  
Order Service: http://localhost:8091  
Kafka UI: http://localhost:8085  

## Local Development (infrastructure in Docker)

```bash
# PostgreSQL and Kafka only
docker compose up -d postgres-user-db postgres-order-db zookeeper kafka

# Run services in separate terminals
go run ./cmd/gateway
go run ./cmd/user-service    # SERVER_PORT=8081
go run ./cmd/user-service    # SERVER_PORT=8082 (second instance)
go run ./cmd/order-service
```

## API Summary (via Gateway)

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /users | Create user (`username`, `initialBalance`) |
| GET | /users/:id | Get user |
| POST | /orders | Create order (`userId`, `amount`) – starts saga |
| GET | /orders/:id | Get order |
| DELETE | /orders/:id | Cancel order (compensation) |

## Saga Flow

1. **POST /orders** → Order service creates order with PENDING and publishes to `order.created`.
2. User service consumes `order.created` and attempts credit reservation.
3. On success: `user.credit-reserved` → Order service sets status to CONFIRMED.
4. On failure: `user.credit-reservation-failed` → Order service sets status to CANCELED.
5. **DELETE /orders/:id** → Order service sets CANCELED, publishes `order.canceled`; user service restores balance (compensation).

## License

MIT
