# Phase 0 — Conventions & Architecture

**Goal:** Establish the rules every later phase relies on — the layering contract, a single JSON response
envelope, a typed error model, and the recipe for adding a feature end-to-end. This phase is mostly
reading + two small shared files; it changes no behavior.

**Depends on:** nothing.
**Risk:** Low. Foundational. Read this once; refer back from every other phase.

---

## Checklist
- [ ] 0.1 Internalize the layering contract (what may import what).
- [ ] 0.2 Add the response envelope helpers (`internal/delivery/httpdelivery/response.go`).
- [ ] 0.3 Add the typed app-error model (`internal/domain/apperror.go`).
- [ ] 0.4 Learn the "add a feature" recipe (domain → repo → usecase → handler → route).
- [ ] 0.5 Note the known bugs to be fixed in Phase 1.

---

## 0.1 The layering contract

```
delivery  ──imports──▶ usecase ──imports──▶ domain ◀──implements── infrastructure
   (gin)                 (logic)              (interfaces + entities)     (pgx, http)
```

Rules:
- `domain` imports **only** the standard library. No gin, no pgx. Entities + interfaces live here.
- `usecase` imports `domain` only. Pure business logic; no `*gin.Context`, no SQL strings.
- `infrastructure` imports `domain` and implements its interfaces (repositories, external clients).
- `delivery` imports `usecase` + `domain`; wires everything in `NewHandler`. Handlers are **thin**.

If you find yourself importing `gin` inside a usecase or writing SQL in a handler, stop — you're crossing a
boundary. The current code already follows this; keep it that way.

> Note: today some handlers call repositories directly (e.g. `problemsLike` calls `h.repos.Problem.Like`).
> That's acceptable for trivial pass-throughs, but anything with business rules (per-user state, streak
> math, recommendations) goes in a usecase. Phases 4, 7, 8 introduce usecases for exactly this reason.

## 0.2 Response envelope

Today handlers return ad-hoc shapes: bare arrays, `gin.H{"error":..., "message":...}`, `gin.H{"success":true}`.
The mobile client tolerates this, but it makes errors inconsistent. Adopt **two** helpers and use them
everywhere from Phase 1 onward. Copy [`response.go`](./response.go) to
`internal/delivery/httpdelivery/response.go`.

- Success that returns data the client parses as a list/object — keep returning the **raw** value
  (`respondOK(c, list)`), because the client's `fromJson` parsers expect the bare entity, not a wrapper.
- Errors — always `respondError(c, status, code, message)` so every failure has the same `{error, message}`
  shape the client already reads.

> Do **not** wrap list/object success responses in `{ "data": ... }` — that would break the existing
> Flutter parsers (`_parseList`, `_asMap`). The envelope standardizes **errors**, and adds optional
> pagination metadata only on the new paginated endpoints (Phase 10).

## 0.3 Typed error model

Copy [`apperror.go`](./apperror.go) to `internal/domain/apperror.go`. It defines `AppError` with a `Code`,
HTTP `Status`, and message, plus sentinel constructors (`ErrNotFound`, `ErrUnauthorized`, `ErrConflict`,
`ErrValidation`, `ErrInternal`). Usecases return `*AppError`; a single helper in `response.go` maps it to
the right HTTP status. This kills the scattered `strings.Contains(err.Error(), "not found")` checks in the
current handler.

## 0.4 The "add a feature" recipe

Every new feature (consistency, learning, bookmarks…) follows the same five steps. Keep them in this order
so the build stays green:

1. **Domain** — `internal/domain/<feature>.go`: entities + a repository interface.
2. **Migration** — `migrations/NNNN_<feature>.up.sql` (+ `.down.sql`) for its tables (Phase 2 onward).
3. **Repository** — `internal/infrastructure/databases/<feature>_repo.go` implements the interface.
4. **Usecase** — `internal/usecase/<feature>/<feature>_usecase.go` for any logic beyond a pass-through.
5. **Delivery** — handler methods + register routes in `routes.go`, wire the repo in `cmd/server/main.go`'s
   `Repos` struct and `NewHandler`.

See [`feature_scaffold.md`](./feature_scaffold.md) for the copy-paste skeleton of all five files.

## 0.5 Known bugs (fixed in Phase 1)

While reading the current code, note these so you recognize them:

1. **HTTP 500 on success.** `problemsLike/Dislike/Solve/Unsolve` end with
   `c.JSON(http.StatusInternalServerError, gin.H{"success": true})` — the success path returns 500. Several
   list handlers do the same in their fallback branch. Fix: return `http.StatusOK`.
2. **`List()` to check existence.** Like/solve handlers call `repos.Problem.List()` and loop to verify the
   id exists before every mutation — an O(n) full scan per write. Replace with `GetById` or rely on
   `RowsAffected() == 0 → not found`.
3. **Daily = `list[0]`.** Not date-based. Fixed in Phase 4.
4. **Global mutable state.** `Like`/`MarkSolved` flip a column on the shared `problems` row. Fixed in Phase 4.

---

## Definition of Done
- [ ] `response.go` and `apperror.go` exist and compile (`go build ./...`).
- [ ] You can articulate the layering rule and the 5-step feature recipe without looking.
- [ ] The four known bugs are noted as Phase-1 work.
