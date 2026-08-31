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

Prerequisites: Go 1.26+, Docker Desktop.

```sh
cp .env.example .env
# set a real JWT_SECRET:  openssl rand -hex 32   (must be >= 32 chars)

docker compose up -d          # postgres + redis + minio
go run ./cmd/api              # applies migrations + seeds catalog on boot, listens on :5000
```

Health check: `curl localhost:5000/healthz` → `{"status":"ok"}`.
Stop infra when done: `docker compose down` (add `-v` to wipe data).

With `make` (optional): `make up`, `make run`, `make test`, `make lint`,
`make build`, `make migrate-create name=add_x`.

### Payments in dev

`LAVA_API_KEY` is blank in `.env.example`, so a **stub payment provider** is
used. Fund a milestone, then confirm the fake payment:

```sh
curl -X POST localhost:5000/api/v1/dev/payments/<paymentId>/pay
```

Set real `LAVA_API_KEY` / `LAVA_WEBHOOK_KEY` / `LAVA_OFFER_ID` to switch to
lava.top; the webhook lands at `POST /api/v1/payments/webhook`.

### Trying the full flow

`scripts/e2e/test_phase*.py` drive the whole marketplace flow end to end with
`urllib` — a good reference for the API. See `scripts/e2e/README.md`.

## Payments

`lava.top` is the payment-acceptance provider only (invoice API + webhooks +
subscriptions). It has no escrow, split payments, or third-party payouts —
escrow is a virtual ledger concept and freelancer payouts run on a separate
rail. See the rework plan, section 6.
