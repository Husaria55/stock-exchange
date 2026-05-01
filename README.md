# Highly-Available Simplified Stock Exchange

A stateless, concurrent-safe stock market simulation built in Go.

## Architecture & Engineering Decisions

**1. Stateless API & Load Balancing**

The Go application holds zero state in memory — everything lives in PostgreSQL. Nginx sits in front of 3 application replicas, distributing traffic and automatically rerouting requests if an instance goes down. This is what makes `POST /chaos` (which kills an instance) a non-event for end users.

**2. Concurrency & Race Condition Prevention**

Running multiple replicas breaks any solution that relies on in-memory locks. Instead, the core trade logic uses ACID transactions with row-level locking (`SELECT ... FOR UPDATE`). When two replicas try to buy the last available stock simultaneously, the database forces them to take turns — preventing double-spending or negative balances entirely.

**3. Why Go?**

Go's minimal memory footprint and near-instant startup mean a killed instance is replaced by Docker in milliseconds, keeping the healthy replica pool intact with barely noticeable degradation.

**4. Clean Architecture**

Three distinct layers keep the codebase testable and maintainable:
- `handler` — parses HTTP requests and writes responses
- `service` — business logic
- `repository` — all database access

## Running the Application

**Prerequisites:** Docker, Docker Compose

```bash
docker-compose up --build
```

This boots Postgres, an Nginx load balancer, and 3 Go API replicas. The API instances wait for Postgres to pass its healthcheck before starting.

Available at: `http://localhost:8000`

## Assumptions & Clarifications

- `POST /stocks` is a hard reset — existing inventory is zeroed before applying the new payload.
- The audit log is capped at 10,000 rows on retrieval, matching the stated operation limit and preventing memory exhaustion.
- Endpoints that return raw numbers (e.g. `GET /wallets/{wallet_id}/stocks/{stock_name}`) respond with `Content-Type: text/plain` to match the spec strictly.

## Testing

```bash
go test -v -race ./...
```

The `-race` flag enables Go's built-in data race detector. An integration test in `internal/repository/postgres_test.go` fires 100 concurrent buy requests against a stock with only 50 units — exactly 50 succeed and 50 fail gracefully, proving the locking mechanism works.

## Pre-submission Checklist

- `go.mod` and `go.sum` are in the root directory
- `go mod tidy` has been run
- Full dry run: delete all containers, run `docker-compose up --build`, and exercise all endpoints via curl or Postman
- Hit `/chaos` and confirm Docker restarts the instance while Nginx handles the next request without interruption