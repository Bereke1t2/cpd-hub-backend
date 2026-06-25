# Phase 12 — Observability & Deploy

**Goal:** Make the running service legible and shippable: structured request logging with the request-id from
Phase 1, Prometheus metrics, a production Docker image (already drafted in Phase 1), and a one-command deploy
config. After this phase you can run it somewhere real and see what it's doing.

**Depends on:** Phases 1 (server/middleware/Dockerfile), 11 (CI green before deploy).
**Risk:** Low. Operational layer; doesn't change API behavior.

---

## Checklist
- [ ] 12.1 Structured JSON request logger (method, path, status, latency, request_id).
- [ ] 12.2 Prometheus `/metrics` (request count + latency histogram by route/status).
- [ ] 12.3 Production env management (`.env.example`, documented vars, secrets out of git).
- [ ] 12.4 Deploy config for a PaaS (Fly.io / Render / Railway) — pick one.
- [ ] 12.5 Run migrations as a deploy step (or boot with `AUTO_MIGRATE`).
- [ ] 12.6 Runbook: how to roll back, read logs, check health.

## 12.1 Structured logging
Replace `gin.Logger()` with a JSON logger that includes the `request_id` set in Phase 1. Copy
[`logger.go`](./logger.go) to `internal/delivery/httpdelivery/logger.go` and add it to the middleware stack.
One structured line per request makes log search/aggregation possible (Grafana Loki, CloudWatch, etc.).

## 12.2 Metrics
Copy [`metrics.go`](./metrics.go) to `internal/delivery/httpdelivery/metrics.go`. It registers a Prometheus
counter (`http_requests_total{route,method,status}`) and a latency histogram, exposes `GET /metrics`, and
provides middleware that records each request. Add `github.com/prometheus/client_golang`. Scrape `/metrics`
from Prometheus; build a dashboard for p50/p95 latency and error rate per route.

## 12.3 Env management
Copy [`env.example`](./env.example) to `.env.example` (commit this; never commit real `.env`). Document every
var: `DATABASE_URL`, `JWT_SECRET`, `SERVER_ADDR`, `APP_ENV`, `CORS_ORIGINS`, `AUTO_MIGRATE`. In production,
set secrets via the platform's secret store, not a file.

## 12.4 Deploy config
Copy [`fly.toml`](./fly.toml) (Fly.io) **or** [`render.yaml`](./render.yaml) (Render) — keep the one you use.
Both reference the Phase-1 multi-stage `Dockerfile`. They define the service, health check (`/healthz`), and
the managed Postgres binding. For Fly: `fly launch` → `fly secrets set JWT_SECRET=... DATABASE_URL=...` →
`fly deploy`.

## 12.5 Migrations on deploy
Two options:
1. **Release step** (preferred for prod): run `go run ./cmd/migrate up` as a deploy/release command so schema
   changes apply before the new code serves traffic. Fly: `[deploy] release_command`. Render: a pre-deploy job.
2. **Boot-apply** behind `AUTO_MIGRATE=true` — simplest, fine for a single instance. Don't auto-migrate with
   multiple instances starting concurrently (migration lock contention).

## 12.6 Runbook
Add `docs/RUNBOOK.md`: how to read logs, the health/ready URLs, how to roll back a deploy, how to run a
migration manually, and what each alert means. Short and practical — the thing you read at 2 AM.

---

## Definition of Done
- [ ] Each request logs one structured JSON line with status, latency, and request_id.
- [ ] `GET /metrics` exposes Prometheus counters + latency histogram.
- [ ] `.env.example` documents every variable; no secrets in git.
- [ ] `fly deploy` (or Render) brings the service up; `/healthz` is green behind the platform's checker.
- [ ] Migrations run as a deploy step or via `AUTO_MIGRATE`; schema is current after deploy.
- [ ] A short RUNBOOK exists.
