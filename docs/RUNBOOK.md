# CPD Hub Backend Runbook

## Health Checks

- Liveness: `GET /healthz` returns `{"status":"ok"}` when the process is up.
- Readiness: `GET /readyz` checks Postgres and returns `503` when the database is unavailable.
- Metrics: `GET /metrics` exposes Prometheus counters and latency histograms.

## Logs

Each request writes one JSON log line with `method`, `path`, `route`, `status`, `bytes`, `latency`, `ip`, and `request_id`.

On Fly.io:

```sh
fly logs -a cpd-hub-backend
```

Use the `request_id` from a response `X-Request-Id` header to find the matching request log.

## Deploy

Set secrets before the first production deploy:

```sh
fly secrets set -a cpd-hub-backend \
  DATABASE_URL='postgres://...' \
  JWT_SECRET="$(openssl rand -hex 32)" \
  CORS_ORIGINS='https://your-web-origin.example'
```

Deploy:

```sh
fly deploy -a cpd-hub-backend
```

The Fly release command runs `/app/migrate -database "$DATABASE_URL" -dir /app/migrations up` before new machines serve traffic.

## Rollback

List releases:

```sh
fly releases -a cpd-hub-backend
```

Rollback to the previous release:

```sh
fly deploy --image "$(fly releases -a cpd-hub-backend --json | jq -r '.[1].image_ref')" -a cpd-hub-backend
```

If the failed release included a schema migration, review the corresponding `migrations/*.down.sql` before rolling the database back.

## Manual Migrations

Run migrations from a local checkout:

```sh
DATABASE_URL='postgres://...' go run ./cmd/migrate -dir migrations up
```

Check migration version:

```sh
DATABASE_URL='postgres://...' go run ./cmd/migrate -dir migrations version
```

For single-instance development environments only, set `AUTO_MIGRATE=true` to apply migrations during server boot.

## Common Alerts

- High `http_request_duration_seconds`: check database readiness, upstream contest fetch logs, and slow endpoints by route.
- Elevated `http_requests_total{status=~"5.."}`: inspect JSON request logs by route and `request_id`.
- `/readyz` failing: verify `DATABASE_URL`, Postgres health, and network reachability from the app.
