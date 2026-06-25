# cpd-hub-backend — Execution Plan & Engineering Playbook

> Go + Gin + PostgreSQL backend for **CPD Hub** (Competitive Programming Division).
> Architecture: Clean / Hexagonal — `domain → usecase → infrastructure → delivery`.
> Module: `github.com/bereket/cpd-hub-backend`. The Flutter client lives in `../CPD_HUB`.

This folder is the **single source of truth** for evolving the backend from a working scaffold into a
production service that fully serves the mobile app. It is split into **phases**. Each phase lives in its
own folder containing:

- `SKILL.md` — goal, audit, step-by-step instructions, checklist, and Definition of Done.
- **template code** (`*.go`, `*.sql`, config files) you copy into the repo and fill in.

Work the phases **in order** — each assumes the previous is merged. Inside a phase, steps are ordered so
`go build ./...` stays green after every step.

---

## 1. Where the backend is today (audit summary)

| Area | State | Notes |
|------|-------|-------|
| Architecture | ✅ Solid | Clean layering: `domain / usecase / infrastructure / delivery`. Interfaces in `domain`. |
| Auth | ✅ Works | `POST /api/auth/{login,signup}` → bcrypt + JWT (`postgres/auth_repo.go`). |
| Problems | ✅ Works | list / daily / by-id / like / dislike / solve / unsolve via Postgres. |
| Contests | ⚠️ Partial | Merges Kontests + Codeforces standings; no caching, no participation, no countdown fields persisted. |
| Profiles / Analytics | ⚠️ Partial | Repo methods exist, but `EnsureAllTables` never creates `submissions`/`rating_history`/`heatmap`/`attendance` tables → those endpoints fail or return nothing. |
| Activity / Info | ⚠️ Static | Listed from DB but never **generated** from user actions. |
| Per-user state | ❌ Missing | `like`/`solve` mutate a **global** row on `problems`. No `loadUser` middleware, no `user_problems` join. Every user sees the same `isLiked`/`solved`. |
| Daily problem | ❌ Naive | Returns `list[0]`, not a date-rotated pick. |
| Identity model | ⚠️ Muddled | `email` is stored as the `username` PK; login queries `username = req.Email`. |
| Success bug | ❌ Bug | `like`/`dislike`/`solve`/`unsolve` and several list fallbacks return **HTTP 500** on the success path (`c.JSON(http.StatusInternalServerError, gin.H{"success": true})`). |
| Migrations | ❌ Ad-hoc | DDL is inline `CREATE TABLE IF NOT EXISTS` in `db.go`. No versioned migrations. |
| Consistency (streaks/goals/ladders) | ❌ None | Mobile feature has **no** backend; client is mock-only. |
| Learning (topics DAG/tracks/lessons) | ❌ None | Mobile feature has **no** backend; client is mock-only. |
| Bookmarks | ❌ None | Mobile has a bookmarks cubit; no endpoint. |
| Validation / pagination / rate-limit / CORS | ❌ None | No request validation, unbounded list responses, no CORS, no throttling. |
| Tests | ❌ None | No `_test.go` files anywhere. |
| Observability / deploy | ❌ None | No structured logging, metrics, health endpoint, graceful shutdown, or Dockerfile (only Postgres `docker-compose`). |

**One-line takeaway:** the architecture and read paths are good; the work is (a) fix correctness bugs,
(b) make state **per-user**, (c) build the three missing feature domains the app already has UIs for
(consistency, learning, bookmarks), and (d) add the production concerns (migrations, validation, tests,
observability, deploy).

---

## 2. Target architecture

```
cmd/
  server/        entrypoint: config → db → migrate → wire repos → http server (graceful)
  seed/          dev seed loader
  migrate/       migration runner CLI                         (NEW, Phase 2)
internal/
  domain/        entities + repository/usecase interfaces     (one file per aggregate)
  usecase/       business logic (no gin, no sql)              (per feature)
  infrastructure/
    config/      typed env config
    postgres/    pgx client + migrations                      (migrations NEW)
    databases/   repository implementations
    external/    kontests / codeforces clients + cache        (cache NEW)
    security/    jwt, password, middleware
  delivery/
    httpdelivery/ gin handlers, routes, middleware, response envelope
migrations/      *.up.sql / *.down.sql                        (NEW, Phase 2)
docs/            this playbook
```

Principles we keep:
- **Dependencies point inward.** `domain` imports nothing from `infrastructure`/`delivery`.
- **Interfaces live in `domain`**, implementations in `infrastructure`.
- **Handlers are thin** — parse → call usecase → shape response. No SQL in handlers.
- **One repository interface per aggregate.**

New rules we add:
- **Every state change is scoped to the authenticated user** (Phase 4).
- **Versioned SQL migrations** are the only way the schema changes (Phase 2).
- **One JSON envelope** for success and error responses (Phase 0 + 10).
- **Usecases are unit-tested; handlers have httptest coverage** (Phase 11).

---

## 3. Phase roadmap

| Phase | Folder | Goal | Depends on |
|-------|--------|------|-----------|
| 0 | [`phase-00-conventions`](./phase-00-conventions/SKILL.md) | Architecture rules, response envelope, error model, how to add a feature, the bugs to fix | — |
| 1 | [`phase-01-foundation-hardening`](./phase-01-foundation-hardening/SKILL.md) | Fix HTTP-500 success bug, typed config, graceful shutdown, CORS, request-id, recovery, `/healthz`, Dockerfile, Makefile | 0 |
| 2 | [`phase-02-database-migrations`](./phase-02-database-migrations/SKILL.md) | golang-migrate, full schema (incl. missing tables), indexes, seed strategy | 1 |
| 3 | [`phase-03-auth-identity`](./phase-03-auth-identity/SKILL.md) | Split username/email, `GET /auth/me`, `loadUser` middleware, refresh tokens, validation | 1, 2 |
| 4 | [`phase-04-user-problem-state`](./phase-04-user-problem-state/SKILL.md) | `user_problems` join → per-user like/dislike/solve, computed counts, date-based daily | 2, 3 |
| 5 | [`phase-05-profiles-analytics`](./phase-05-profiles-analytics/SKILL.md) | Real heatmap / rating-history / attendance / submissions / profile aggregates | 2, 4 |
| 6 | [`phase-06-contests`](./phase-06-contests/SKILL.md) | Cached Kontests/CF client, countdown fields, participation, refresh worker | 2, 3 |
| 7 | [`phase-07-consistency`](./phase-07-consistency/SKILL.md) | **NEW**: streaks, weekly goals, rating ladders (serves mobile consistency feature) | 4 |
| 8 | [`phase-08-learning`](./phase-08-learning/SKILL.md) | **NEW**: topic DAG, tracks, lessons, skill-tree (serves mobile learning feature) | 4 |
| 9 | [`phase-09-activity-feed`](./phase-09-activity-feed/SKILL.md) | Generate activity from real actions, bookmarks, info CMS, pagination | 4, 5 |
| 10 | [`phase-10-api-hardening`](./phase-10-api-hardening/SKILL.md) | Validation, pagination, rate-limit, CORS finalize, OpenAPI spec | all feature phases |
| 11 | [`phase-11-testing-cicd`](./phase-11-testing-cicd/SKILL.md) | Usecase/handler/repo tests, GitHub Actions CI, lint | 1–10 |
| 12 | [`phase-12-observability-deploy`](./phase-12-observability-deploy/SKILL.md) | Structured logs, metrics, prod Dockerfile, deploy config | 1, 11 |
| 13 | [`phase-13-courses`](./phase-13-courses/SKILL.md) | **NEW**: structured courses (modules/lessons), per-user lesson completion (`api.md` §8) | 3, 4 |
| 14 | [`phase-14-smart-practice`](./phase-14-smart-practice/SKILL.md) | **NEW**: SM-2 spaced-repetition review queue + contest upsolves (`api.md` §9) | 3, 4 |
| 15 | [`phase-15-articles-feed`](./phase-15-articles-feed/SKILL.md) | **NEW**: first-party articles feed with filters + pagination (`api.md` §10) | 2, 10 |

**Phases 13–15** were added after `api.md` was extended with the Courses, Smart Practice, and Articles
contracts. They are independent feature domains: each can be built any time after Phase 4 (13, 14) or
Phase 10 (15), and they don't block one another. The Flutter screens (`../CPD_HUB/lib/features/{courses,
practice,articles}`) already exist and run on mock data today.

**Critical path for the mobile app:** Phases 1 → 2 → 3 → 4 unblock correct per-user behavior. Then
7 and 8 light up the consistency and learning screens that currently run on mock data. 5, 6, 9 fill in
profile analytics, contests, and the feed. 10–12 are production-readiness and can overlap once features land.

Phases 5, 6 can run in parallel after 4. Tests should be written incrementally **inside** each phase, not
deferred — Phase 11 only adds CI + fills gaps.

---

## 4. How to use these docs

1. Open the phase folder's `SKILL.md`. Read **Goal** and **Definition of Done**.
2. Copy the template files from the same folder into the path named in each template's header comment.
3. Work the checklist top-to-bottom; keep `go build ./...` and `go vet ./...` clean after each step.
4. Tick the Definition of Done before moving on.

### Standing commands
```bash
go mod tidy
go build ./...                      # must stay green
go vet ./...
go test ./...                       # green before merge (from Phase 11)
gofmt -l .                          # empty output before commit
DATABASE_URL=postgres://postgres:postgres@localhost:5432/cpdhub go run ./cmd/server
```

---

## 5. The API contract (what the mobile app expects)

The Flutter client's paths are pinned in `../CPD_HUB/lib/core/url_constants.dart`. Keep these stable —
changing a path or a JSON field name is a breaking change that must be coordinated with the client.

Existing contract (already served):

```
POST   /api/auth/login                                  → { token, user:{username, fullName} }
POST   /api/auth/signup                                 → { token, user:{username, fullName} }
GET    /api/problems                                    → [Problem]
GET    /api/problems/daily                              → Problem
GET    /api/problems/:id                                → Problem
POST   /api/problems/:id/like | /dislike | /solve       → { success }
DELETE /api/problems/:id/solve                          → { success }
GET    /api/contests                                    → [Contest]
GET    /api/contests/:id/leaderboard                    → [LeaderboardEntry]
GET    /api/users                                       → [UserProfile]
GET    /api/users/profile/:username                     → UserProfile
GET    /api/users/profile/:username/analytics/heatmap   → [HeatmapEntry]
GET    /api/users/profile/:username/analytics/rating-history → [RatingEntry]
GET    /api/users/profile/:username/attendance          → [AttendanceEntry]
GET    /api/users/profile/:username/submissions         → [Submission]
GET    /api/activity                                    → [Activity]
GET    /api/info                                        → [Info]
```

New endpoints these phases add (mobile already has the screens):

```
GET    /api/auth/me                                     → UserProfile                  (Phase 3)
POST   /api/auth/refresh                                → { token }                    (Phase 3)
GET    /api/consistency/streak   | PUT                  → Streak                       (Phase 7)
GET    /api/consistency/goal     | PUT                  → Goal                         (Phase 7)
GET    /api/consistency/ladders                         → [Ladder]                     (Phase 7)
GET    /api/learning/topics                             → [Topic]    (DAG)             (Phase 8)
GET    /api/learning/tracks                             → [Track]                      (Phase 8)
GET    /api/learning/lessons/:topicId                   → Lesson                       (Phase 8)
GET    /api/bookmarks  | POST /:problemId | DELETE /:problemId                         (Phase 9)
GET    /api/courses    | GET /:id                        → [Course] / Course           (Phase 13)
POST   /api/courses/:courseId/lessons/:lessonId/complete → { lessonId, completed }      (Phase 13)
GET    /api/practice/review-queue | POST | PUT /:problemId | DELETE /:problemId         (Phase 14)
GET    /api/practice/upsolves     | POST | PUT /:problemId                              (Phase 14)
GET    /api/articles?limit&offset&source&tag            → [Article]                    (Phase 15)
```

> When you add or change an endpoint, update **both** `routes.go` and the OpenAPI spec (Phase 10), and
> grep the Flutter `url_constants.dart` to confirm the path/field names match before you call it done.

---

## 6. Risk notes
- **Per-user migration (Phase 4) is the riskiest change.** Today `solved`/`is_liked` are columns on
  `problems`. Moving them to a `user_problems` join changes every problem read and write. Do it behind the
  existing handler shape so the client JSON (`isLiked`, `solved`) is unchanged.
- **Identity (Phase 3):** the client stores `username` from the token and builds `/users/profile/:username`.
  If you split email from username, keep the token's `username` claim authoritative and backfill existing rows.
- **External APIs (Phase 6):** Kontests/Codeforces are rate-limited and occasionally down. Always cache and
  degrade gracefully — never let a contests fetch failure 500 the whole screen.
- **Secrets:** `JWT_SECRET` must be set in production (the dev fallback is insecure). Never commit `.env`.
