# Phase 1 — Foundation Hardening

**Goal:** Fix the correctness bugs found in Phase 0, then make the process production-shaped: typed config,
graceful shutdown, panic recovery, CORS, request IDs, a `/healthz` endpoint, a Dockerfile, and a Makefile.
No new features — this is about a server that starts cleanly, stops cleanly, and never returns 500 on success.

**Depends on:** Phase 0 (`response.go`, `apperror.go`).
**Risk:** Low–Medium. Touches `main.go` and the handler, but behavior only gets *more* correct.

---

## Checklist
- [ ] 1.1 Fix the HTTP-500-on-success bug in problem write handlers + list fallbacks.
- [ ] 1.2 Replace `List()`-to-check-existence with `RowsAffected`/`GetById`.
- [ ] 1.3 Typed config struct with validation (`infrastructure/config/config.go`).
- [ ] 1.4 Graceful shutdown + timeouts (`cmd/server/main.go` via `server.go` helper).
- [ ] 1.5 Middleware: recovery, request-id, CORS (`delivery/httpdelivery/middleware.go`).
- [ ] 1.6 `GET /healthz` + `GET /readyz` (DB ping).
- [ ] 1.7 Dockerfile (multi-stage) + `.dockerignore`.
- [ ] 1.8 Makefile with the standing commands.

---

## 1.1 Fix the success bug
In `handler.go`, every problem write handler ends like this:

```go
	c.JSON(http.StatusInternalServerError, gin.H{"success": true}) // BUG: 500 on success
```

Replace the trailing line in `problemsLike`, `problemsDislike`, `problemsSolve`, `problemsUnsolve` with:

```go
	respondSuccess(c) // 200 {success:true}
```

Do the same audit for the list handlers: `problemsList`, `problemsDaily`, `activityList`, `infoList`, and
`contestLeaderboard` all have a fallback branch that uses `http.StatusInternalServerError` for what is
really "no repo wired" sample data. Either delete the sample fallback (repos are always wired now) or return
`http.StatusOK`. Prefer deleting — the server requires a DB to boot, so the nil-repo branches are dead code.

## 1.2 Stop scanning to check existence
Handlers call `repos.Problem.List()` and loop to confirm the id exists before each like/solve. Delete that
loop. The repository methods already return `fmt.Errorf("not found")` when `RowsAffected() == 0`. Map that to
a 404 in the handler:

```go
func (h *handlerImpl) problemsLike(c *gin.Context) {
	id := c.Param("id")
	if err := h.repos.Problem.Like(id); err != nil {
		respondError(c, err) // becomes the typed error once repos return *domain.AppError (Phase 4)
		return
	}
	respondSuccess(c)
}
```

While here, change the repo's `fmt.Errorf("not found")` to `domain.ErrNotFound("problem not found")` so
`respondError` produces a 404 instead of a 500.

## 1.3 Typed config
The current `config.Load()` reads a couple of env vars. Replace with a validated struct — copy
[`config.go`](./config.go) to `internal/infrastructure/config/config.go`. It fails fast with a clear message
if `DATABASE_URL` is missing or `JWT_SECRET` is the insecure default in a non-dev environment.

## 1.4 Graceful shutdown
`main.go` calls `srv.ListenAndServe()` and never handles SIGTERM, so in-flight requests are killed on deploy.
Copy [`server.go`](./server.go) to `internal/delivery/httpdelivery/server.go` and use it from `main.go`:

```go
srv := httpdelivery.NewServer(cfg.Server.Address, h.Router())
if err := srv.Run(ctx); err != nil { // blocks; returns on SIGINT/SIGTERM
	log.Fatalf("server error: %v", err)
}
```

It sets read/write/idle timeouts and drains connections with a 10s deadline on shutdown.

## 1.5 Middleware
Copy [`middleware.go`](./middleware.go) to `internal/delivery/httpdelivery/middleware.go`. It provides:
- `RecoveryJSON()` — converts panics into a 500 JSON error instead of crashing the worker.
- `RequestID()` — attaches/propagates `X-Request-Id`.
- `CORS(allowedOrigins)` — needed for the Flutter **web** build and any browser client. For mobile it's
  harmless. Read origins from config (`CORS_ORIGINS`, comma-separated; default `*` in dev).

Apply them in `NewHandler` **before** `RegisterRoutes`:

```go
g := gin.New() // not gin.Default(); we add our own logger + recovery
g.Use(RequestID(), RecoveryJSON(), CORS(cfg.CORSOrigins))
```

> Switch `gin.Default()` → `gin.New()` so you control the middleware stack. Add a JSON request logger in
> Phase 12; for now `gin.Logger()` is fine.

## 1.6 Health endpoints
Copy [`health.go`](./health.go) to `internal/delivery/httpdelivery/health.go`. Register **outside** the
`/api` auth group so probes don't need a token:

```go
r.GET("/healthz", h.Healthz)   // liveness: always 200 if process is up
r.GET("/readyz", h.Readyz)     // readiness: 200 only if DB ping succeeds
```

Pass the `*postgres.Client` into the handler so `Readyz` can `Pool.Ping(ctx)`.

## 1.7 Dockerfile
Copy [`Dockerfile`](./Dockerfile) and [`.dockerignore`](./dockerignore.txt) (rename to `.dockerignore`) to
the repo root. Multi-stage: build a static binary on `golang:1.22`, run on `gcr.io/distroless/static`. Final
image is ~15 MB and runs as non-root.

## 1.8 Makefile
Copy [`Makefile`](./Makefile) to the repo root. Targets: `run`, `build`, `test`, `vet`, `fmt`, `tidy`,
`docker-build`, `migrate-up`, `migrate-down`, `seed`.

---

## Definition of Done
- [ ] A successful like/dislike/solve returns **200**, not 500.
- [ ] Unknown problem id returns **404** with `{error:"not_found"}`.
- [ ] `Ctrl-C` / SIGTERM drains in-flight requests then exits within 10s.
- [ ] A panic in a handler returns a 500 JSON body; the server stays up.
- [ ] `GET /healthz` → 200 always; `GET /readyz` → 200 with DB, 503 without.
- [ ] `docker build .` produces a runnable image; `make run` boots the server.
- [ ] `go build ./...`, `go vet ./...`, `gofmt -l .` all clean.
