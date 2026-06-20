# Subscriptions Aggregation Service

A REST service for aggregating data about users' online subscriptions.
It provides CRUDL operations over subscriptions and calculates the total cost of subscriptions for a selected period.

## Features

- CRUDL over subscription records (`service_name`, `price`, `user_id`, `start_date`, optional `end_date`).
- Total cost calculation for a period with filtering by `user_id` and `service_name`.
- PostgreSQL + migrations (applied automatically on service startup).
- Structured logging (`log/slog`).
- Configuration via `.env` / environment variables.
- Swagger documentation (OpenAPI 3.0).
- One-command startup via `docker compose`.

## Tech stack

- Go 1.26, [chi](https://github.com/go-chi/chi) router
- PostgreSQL 16, [pgx/v5](https://github.com/jackc/pgx) driver
- [golang-migrate](https://github.com/golang-migrate/migrate) migrations (embedded into the binary via `embed`)
- Swagger UI via [http-swagger](https://github.com/swaggo/http-swagger)

## Project layout

```
cmd/server/          — entry point (main)
internal/config/     — configuration loading
internal/logger/     — slog setup
internal/model/      — domain entity, DTOs, MM-YYYY date type
internal/repository/ — data access (SQL)
internal/service/    — business logic
internal/handler/    — HTTP handlers
internal/router/     — router assembly, middleware, swagger
internal/database/   — connection pool and migrations
migrations/          — SQL migrations (embedded into the binary)
docs/                — OpenAPI specification
```

Layers: `handler → service → repository`. Dependencies point inward, and layers communicate through interfaces (which makes testing easy).

## Quick start (docker compose)

```bash
cp .env.example .env      # adjust values if needed
docker compose up --build
```

Two containers start: `db` (PostgreSQL) and `app` (the service). Migrations are applied automatically.

- API: `http://localhost:8080/api/v1`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Health check: `http://localhost:8080/health`

## Local run (without Docker)

Requires a running PostgreSQL. Connection parameters are in `.env`.

```bash
go mod download
go run ./cmd/server
```

## Configuration

All parameters are read from the environment (for local runs they are also picked up from `.env`).
Real environment variables take precedence over `.env`, which is convenient for docker-compose.

| Variable                | Default         | Description                       |
|-------------------------|-----------------|-----------------------------------|
| `HTTP_PORT`             | `8080`          | HTTP server port                  |
| `HTTP_READ_TIMEOUT`     | `10s`           | Request read timeout              |
| `HTTP_WRITE_TIMEOUT`    | `10s`           | Response write timeout            |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s`           | Graceful shutdown timeout         |
| `DB_HOST`               | `localhost`     | PostgreSQL host                   |
| `DB_PORT`               | `5432`          | PostgreSQL port                   |
| `DB_USER`               | `postgres`      | User                              |
| `DB_PASSWORD`           | `postgres`      | Password                          |
| `DB_NAME`               | `subscriptions` | Database name                     |
| `DB_SSLMODE`            | `disable`       | SSL mode                          |
| `LOG_LEVEL`             | `info`          | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT`            | `text`          | `text` / `json`                   |

## API

Base prefix: `/api/v1`.

| Method | Path                       | Description                           |
|--------|----------------------------|---------------------------------------|
| POST   | `/subscriptions`           | Create a subscription                 |
| GET    | `/subscriptions/{id}`      | Get a subscription                    |
| PUT    | `/subscriptions/{id}`      | Update a subscription                 |
| DELETE | `/subscriptions/{id}`      | Delete a subscription                 |
| GET    | `/subscriptions`           | List subscriptions (filters + paging) |
| GET    | `/subscriptions/summary`   | Total cost for a period               |

Start/end dates are passed in `MM-YYYY` format (e.g. `07-2025`).

### Request examples

Create a subscription:

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

List with a filter:

```bash
curl 'http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&limit=20'
```

Total cost for a period (`user_id` and `service_name` filters are optional):

```bash
curl 'http://localhost:8080/api/v1/subscriptions/summary?from=01-2025&to=12-2025&service_name=Yandex%20Plus'
```

Response:

```json
{ "total_price": 2400 }
```

### How the total cost is calculated

The subscription price is per month, so the cost of a single subscription over a period `[from, to]` =
`price × (number of months its active range overlaps the period)`.
Subscriptions without an `end_date` are treated as active until the end of the requested period.
The result is the sum over all subscriptions matching the filter.

## Migrations

SQL files live in `migrations/` and are embedded into the binary (`go:embed`).
They are applied automatically on service startup (`migrate ... Up()`); re-running is idempotent.

## Tests

```bash
go test ./...
```

Covered: parsing/validation of the `MM-YYYY` date format, input validation, and the HTTP handlers
(via `httptest` with a stub service, so no database is required).
