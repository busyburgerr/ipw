# ipw backend

Go backend for the freelance marketplace (rewritten from the "IT Professionals
Work" job board — see the rework plan).

## Stack

- Go 1.26 · Fiber v2 · GORM · PostgreSQL 16 · Redis 7
- JWT access tokens + hashed refresh-token sessions
- `golang-migrate` (SQL migrations embedded in the binary)
- Object storage: S3-compatible (MinIO locally)

## Layout

```
cmd/api/                 composition root / HTTP entrypoint
internal/
  config/                env-based configuration
  httpx/                 Fiber server, error envelope, middleware
  platform/postgres/     DB connection + embedded migrations
  platform/redis/        Redis connection
  user/                  account entity + persistence contract
  auth/                  register / login / refresh / logout, guards
```

Each feature package owns its domain type, its GORM row mapping, its service and
its HTTP handler. Domain types carry no ORM tags; handlers return DTOs, never
GORM models.

## Local development

```sh
cp .env.example .env          # then set JWT_SECRET (openssl rand -hex 32)
make up                       # postgres + redis + minio via docker compose
make run                      # applies migrations on boot, listens on :5000
```

Other targets: `make test`, `make lint`, `make build`, `make migrate-create name=add_projects`.

## Payments

`lava.top` is the payment-acceptance provider only (invoice API + webhooks +
subscriptions). It has no escrow, split payments, or third-party payouts —
escrow is a virtual ledger concept and freelancer payouts run on a separate
rail. See the rework plan, section 6.
