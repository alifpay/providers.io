# providers.io

A demo payment service where agents submit payments via HTTP API. Payments are persisted immediately, then processed asynchronously via a **Temporal** workflow with automatic retries.

## Architecture

```
Agent → POST /payments → Insert (pending) → 202 Accepted
                                  ↓
                        Temporal Workflow
                                  ↓
                        CallProviderActivity  ← retries on transient errors
                                  ↓
                        UpdatePaymentStatusActivity
                                  ↓ (on success)
                        InsertBalanceLogActivity
                        UpsertDailyStatsActivity
```

Two binaries:
- **`cmd/server`** — HTTP API, accepts requests and starts workflows
- **`cmd/worker`** — Temporal worker, executes workflow activities

## Project Structure

```
internal/
  config/         — env-based config
  domain/         — Payment, Provider, Partner, BalanceLog, DailyStats types
  repository/     — pgx implementations (payments, balance_log, daily_stats, providers)
  provider/       — mock provider client (70% success / 20% retryable / 10% terminal)
  workflow/       — Temporal workflow + activities
  api/            — HTTP handlers, router, middleware
cmd/
  server/         — HTTP server entry point
  worker/         — Temporal worker entry point
docs/
  db.sql          — TimescaleDB schema
```

## Database Schema (TimescaleDB)

**Write side:**
- `partners` — partner accounts
- `providers` — payment provider connections (per partner)
- `payments` — all payment records (hypertable, partitioned by `pay_date`)
- `balance_log` — immutable audit log of every balance movement (hypertable)

**Read side (projections):**
- `daily_stats_projection` — aggregated counts and amounts per provider per day (hypertable)

## Payment Workflow

1. Agent calls `POST /payments`
2. Payment inserted with `status=1` (pending)
3. `202 Accepted` returned immediately with `payment_id`
4. Temporal workflow starts asynchronously:
   - Calls mock provider with retry (up to 10 attempts, exponential backoff 1s→30s)
   - Terminal error codes (e.g. insufficient funds) stop retries immediately
   - Updates payment status to `2` (success) or `3` (failed)
   - On success: inserts to `balance_log` and upserts `daily_stats_projection`

## Payment Status Codes

| Value | Meaning  |
|-------|----------|
| 1     | Pending  |
| 2     | Success  |
| 3     | Failed   |
| 4     | Canceled |

## Provider Error Codes

| Code | Type      | Description        |
|------|-----------|--------------------|
| 0    | —         | None               |
| 101  | Terminal  | Insufficient funds |
| 102  | Terminal  | Invalid account    |
| 201  | Retryable | Timeout            |
| 202  | Retryable | Provider down      |

## Running Locally

**Prerequisites:** Go 1.22+, PostgreSQL with TimescaleDB, Temporal CLI

```bash
# 1. Apply schema
psql $PGURL -f docs/db.sql

# 2. Start Temporal dev server
temporal server start-dev

# 3. Start worker
PGURL=postgres://user:pass@localhost:5432/db go run ./cmd/worker

# 4. Start API server
PGURL=postgres://user:pass@localhost:5432/db go run ./cmd/server

# 5. Create a payment
curl -X POST localhost:8080/payments \
  -H 'Content-Type: application/json' \
  -d '{"provider_id":1,"agent_id":"agent1","ref_id":"ref-001","amount":"100.50"}'
# → {"payment_id":"...","status":"pending"}

# 6. Poll status
curl localhost:8080/payments/{payment_id}

# 7. View workflow in Temporal UI
open http://localhost:8233
```

## Environment Variables

| Variable            | Default         | Description               |
|---------------------|-----------------|---------------------------|
| `PGURL`             | required        | PostgreSQL connection URL |
| `TEMPORAL_HOST`     | `localhost:7233`| Temporal server address   |
| `TEMPORAL_NAMESPACE`| `default`       | Temporal namespace        |
| `HTTP_ADDR`         | `:8080`         | HTTP listen address       |

## Idempotency

- DB unique index on `(agent_id, ref_id, pay_date)` — duplicate inserts return `409 Conflict`
- Temporal `WorkflowID = payment_id` — duplicate workflow starts are a no-op
