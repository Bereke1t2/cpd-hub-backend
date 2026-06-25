# Phase 2 — Database & Migrations

**Goal:** Replace the ad-hoc `CREATE TABLE IF NOT EXISTS` block in `postgres/db.go` with **versioned
migrations**, and add the tables later phases need but `EnsureAllTables` never created (`submissions`,
`rating_history`, `daily_solves`/heatmap, `attendance`, plus the join/feature tables). One source of truth
for the schema, reproducible from zero.

**Depends on:** Phase 1.
**Risk:** Medium. Schema is the foundation everything else reads/writes. Do it before the per-user work.

---

## Checklist
- [x] 2.1 Add `golang-migrate` (library + a tiny `cmd/migrate` CLI).
- [x] 2.2 Author the baseline migration `0001_init` (existing tables, cleaned up).
- [x] 2.3 Author `0002_analytics` (submissions, rating_history, daily_solves, attendance).
- [x] 2.4 Author `0003_user_problems` (per-user join — used by Phase 4).
- [x] 2.5 Author `0004_features` (bookmarks; consistency + learning come with their phases).
- [x] 2.6 Run migrations on boot **or** via CLI; remove `EnsureAllTables`.
- [x] 2.7 Indexes on every foreign key + hot query column.
- [x] 2.8 Rework the seed to insert through the migrated schema.

---

## 2.1 Tooling
Add the migration library:
```bash
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
```
Copy [`migrate_main.go`](./migrate_main.go) to `cmd/migrate/main.go` — a CLI wrapping up/down/version, used
by the Makefile (`make migrate-up`). Copy [`runner.go`](./runner.go) to
`internal/infrastructure/postgres/migrate.go` so the server can also auto-apply migrations on boot in dev.

Migrations live in `./migrations` as paired files: `NNNN_name.up.sql` / `NNNN_name.down.sql`.

## 2.2–2.5 The migration set
Copy the SQL templates from this folder to `./migrations/`:

| File | Creates |
|------|---------|
| [`0001_init.up.sql`](./0001_init.up.sql) / `.down.sql` | `users`, `problems`, `contests`, `profiles`, `activity`, `info` (the current tables, with `created_at`/`updated_at` and proper types) |
| [`0002_analytics.up.sql`](./0002_analytics.up.sql) / `.down.sql` | `submissions`, `rating_history`, `daily_solves` (heatmap source), `attendance` |
| [`0003_user_problems.up.sql`](./0003_user_problems.up.sql) / `.down.sql` | `user_problems` join (per-user like/dislike/solve) |
| [`0004_bookmarks.up.sql`](./0004_bookmarks.up.sql) / `.down.sql` | `bookmarks` |

> Phases 7 (consistency) and 8 (learning) ship their **own** migrations (`0005_*`, `0006_*`) inside those
> phase folders, following the same numbering. Keep numbers globally monotonic.

Key schema decisions baked into these files:
- **`users.username`** is the stable identity (PK). `users.email` becomes a separate **unique** column
  (Phase 3 stops overloading username with the email). `0001` adds `email` as nullable + unique; Phase 3
  backfills.
- **Per-user state** (`user_problems`): `(username, problem_id)` composite PK, columns `liked`, `disliked`,
  `solved`, `solved_at`. The `problems` table keeps `likes`/`dislikes` only as **denormalized counters**
  (or drop them and compute — see Phase 4 trade-off note).
- **Heatmap** is derived from `daily_solves(username, day, count)` — one row per user per day. Cheap to query
  for the calendar UI.
- All timestamps are `TIMESTAMPTZ`. All money/score ints. No `TEXT` timestamps (the current `activity.timestamp`
  is `TEXT` — `0001` keeps it for client-compat but adds a real `created_at TIMESTAMPTZ`).

## 2.6 Apply on boot, drop EnsureAllTables
In `cmd/server/main.go`, after connecting:
```go
if err := postgres.RunMigrations(cfg.Database.URL, "migrations"); err != nil {
	log.Fatalf("migrations failed: %v", err)
}
```
Delete `EnsureAllTables` and its call. In production you may prefer running `make migrate-up` as a separate
deploy step instead of on boot — keep the boot-apply behind an `AUTO_MIGRATE` env flag.

## 2.7 Indexes
Every FK and every column used in a `WHERE`/`ORDER BY` gets an index. The templates include:
`idx_user_problems_username`, `idx_submissions_username_time`, `idx_daily_solves_username_day`,
`idx_activity_created_at`, `idx_contests_start_time`. Add more as queries appear; never ship a new query
that scans without one on the hot path.

## 2.8 Seed
The existing `cmd/seed` + `seed.sql` predate this schema. Update them to match the migrated columns. Keep the
rule from the README: **never insert plaintext passwords** — create users through signup. Seed problems,
contests, info, topic graph (Phase 8), and a couple of `daily_solves` rows so the heatmap renders in dev.

---

## Definition of Done
- [x] `make migrate-up` on an empty database produces the full schema; `make migrate-down` reverses it.
- [x] `EnsureAllTables` is deleted; the server boots and serves against the migrated schema.
- [x] `submissions`, `rating_history`, `daily_solves`, `attendance`, `user_problems`, `bookmarks` all exist.
- [x] Every FK has an index.
- [x] `cmd/seed` populates a working dev dataset against the new schema.
