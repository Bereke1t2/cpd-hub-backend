# Phase 10 — API Hardening

**Goal:** Make the API production-grade at the edges: validate every input, paginate every list, rate-limit
auth and writes, finalize CORS, and publish an OpenAPI spec so the contract is documented and testable.

**Depends on:** the feature phases (1–9).
**Risk:** Low. Additive guards; don't change response shapes the client parses.

---

## Checklist
- [ ] 10.1 Validate all POST/PUT bodies (extend the Phase-3 `bindJSON` to every write).
- [ ] 10.2 Pagination helper for list endpoints (`?limit=&offset=`), sane caps.
- [ ] 10.3 Rate limiting on `/auth/*` and write actions (token-bucket per IP/user).
- [ ] 10.4 CORS finalized from config; lock down origins in production.
- [ ] 10.5 Security headers + body size limit.
- [ ] 10.6 OpenAPI 3 spec (`docs/openapi.yaml`) covering every route.

## 10.1 Validation everywhere
Every handler that binds a body uses `bindJSON` (Phase 3) so malformed/invalid input returns a clean 400.
Add `binding` tags to the remaining request structs (goal, streak, etc.). Reject unknown problem/contest ids
with 404, not 500.

## 10.2 Pagination
Copy [`pagination.go`](./pagination.go) to `internal/delivery/httpdelivery/pagination.go`. It parses
`limit`/`offset` with defaults (20/0) and a max (100). Apply to `GET /problems`, `/users`, `/activity`,
`/submissions`. Pass the bounds into the repo `List` queries (`LIMIT $n OFFSET $m`). Keep returning a bare
array unless you choose the metadata envelope — if you add `{items, total, limit, offset}`, update the client
parser in the same PR.

## 10.3 Rate limiting
Copy [`ratelimit.go`](./ratelimit.go) to `internal/delivery/httpdelivery/ratelimit.go`: an in-memory
token-bucket keyed by client IP (and by username for authenticated writes). Apply a strict bucket to
`/auth/login` and `/auth/signup` (e.g. 5/min) to blunt credential stuffing, and a looser one to write actions.
For multi-instance deployments, swap the in-memory store for Redis later — the interface stays the same.

## 10.4 CORS
The Phase-1 `CORS` middleware reads origins from config. In production set `CORS_ORIGINS` to the explicit web
origin(s); never ship `*` with credentials. Mobile native clients don't send `Origin`, so they're unaffected.

## 10.5 Security headers + body cap
Add a small middleware setting `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, and a
`http.MaxBytesReader` (e.g. 1 MB) on request bodies so a giant payload can't exhaust memory. Include it in the
middleware stack from Phase 1.

## 10.6 OpenAPI spec
Copy [`openapi.yaml`](./openapi.yaml) to `docs/openapi.yaml` and complete it: every path, request body, and
response schema. Keep it in sync with `routes.go` — a stale spec is worse than none. Optionally serve it at
`GET /openapi.yaml` and a Swagger UI at `/docs` (dev only). This is the artifact the mobile team checks field
names against.

---

## Definition of Done
- [ ] Every write endpoint rejects invalid bodies with a 400 + clear message.
- [ ] List endpoints honor `limit`/`offset` and cap at 100.
- [ ] Auth endpoints are rate-limited; rapid login attempts get 429.
- [ ] Production CORS allows only the configured origins.
- [ ] Oversized bodies are rejected; security headers present.
- [ ] `docs/openapi.yaml` describes every route and matches the running server.
