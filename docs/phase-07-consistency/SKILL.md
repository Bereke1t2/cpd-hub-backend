# Phase 7 — Consistency Engine (NEW feature)

**Goal:** Build the backend for the mobile **consistency** feature (streaks, weekly goals, rating ladders)
which today runs entirely on a mock data source. Persist per-user streak/goal state and serve the rung
ladders, so progress survives reinstalls and syncs across devices.

**Depends on:** Phase 4 (solves drive the streak; `daily_solves` already exists).
**Risk:** Medium. Net-new domain, but it reuses signals (`daily_solves`, `user_problems.solved`) already
captured. Follows the Phase-0 feature recipe exactly.

---

## The contract (must match the Flutter models — snake_case)
From `../CPD_HUB/lib/features/consistency/data/models/`:

```
Streak  { current:int, longest:int, last_active_day:date?, freezes_available:int(=2), active_days:[date] }
Goal    { id:str(=weekly-problems), type:str(=problemsPerWeek), target:int(=5), progress:int, period_start:date }
Ladder  { id, title, from_rating:int, to_rating:int, rungs:[ {problem_id, rating, solved, topic_id?} ] }
```

Endpoints:
```
GET /api/consistency/streak              → Streak
PUT /api/consistency/streak   body Streak → Streak
GET /api/consistency/goal                → Goal
PUT /api/consistency/goal     body Goal   → Goal
GET /api/consistency/ladders             → [Ladder]   (base ladders + caller's solved overlay)
PUT /api/consistency/ladders/:id body Ladder → Ladder  (saveLadder override; optional)
```

---

## Checklist
- [x] 7.1 Domain entities + repository interface (`internal/domain/consistency.go`).
- [x] 7.2 Migration `0008_consistency` (`streaks`, `goals`, `ladders` + `ladder_rungs`, `user_ladder_solved`).
- [x] 7.3 Repository (`consistency_repo.go`).
- [x] 7.4 Usecase: derive/recompute streak from `daily_solves`; default goal for new users.
- [x] 7.5 Handlers + routes + wire repo in `main.go`.
- [x] 7.6 Seed the base ladders (Div 2 A/B ladders, etc.).

## 7.1 Domain
Copy [`consistency.go`](./consistency.go) to `internal/domain/consistency.go`. JSON tags are snake_case to
match the client. Note `Streak.LastActiveDay` is a nullable date → use `*string` (formatted `YYYY-MM-DD`) or
`*time.Time` with a custom date format; the template uses formatted strings to avoid timezone surprises.

## 7.2 Migration
Copy [`0008_consistency.up.sql`](./0008_consistency.up.sql) / `.down.sql` to `migrations/`. Tables:
- `streaks(username PK, current, longest, last_active_day, freezes_available)`.
- `streak_active_days(username, day)` — the `active_days` set.
- `goals(username, id, type, target, progress, period_start)` — PK `(username, id)`.
- `ladders(id PK, title, from_rating, to_rating)` + `ladder_rungs(ladder_id, problem_id, rating, topic_id, ord)`.
- `user_ladder_solved(username, problem_id)` — the per-user solved overlay for ladder rungs.

## 7.3 Repository
Copy [`consistency_repo.go`](./consistency_repo.go). `GetLadders(username)` loads base ladders + rungs, then
overlays `solved` from `user_ladder_solved` (or reuse `user_problems.solved` keyed by `problem_id` so solving
a problem anywhere lights up the ladder — preferred; the template does this).

## 7.4 Usecase — streak recompute
The streak is **derived**, not just stored. Copy [`consistency_usecase.go`](./consistency_usecase.go). On
`GET /streak` it recomputes from `daily_solves`:
- `current` = length of the consecutive-day run ending today (or yesterday, with a freeze).
- `longest` = max run ever.
- `active_days` = the set of days with ≥1 solve.
- A **freeze** lets one gap day not break the streak; decrement `freezes_available` when used.

This keeps the streak honest even if the client never PUTs. `PUT /streak` is still accepted (e.g. to spend a
freeze or for offline reconciliation) but the server is the source of truth. The goal's `progress` is
likewise recomputed from solves within `[period_start, period_start+7d)`.

## 7.5–7.6 Delivery + seed
Add `GetStreak/PutStreak/GetGoal/PutGoal/GetLadders/PutLadder` handlers, register routes under a
`/consistency` protected group, add `Consistency domain.ConsistencyRepository` to `Repos` + the `Handler`
interface, wire in `main.go`. Seed a couple of base ladders in `cmd/seed` so `GET /ladders` returns content.

---

## Definition of Done
- [x] `GET /consistency/streak` returns a streak computed from the caller's real solves; solving today bumps
      `current`.
- [x] A new user gets a sensible default goal (`weekly-problems`, target 5) without a 404.
- [x] `GET /consistency/goal` shows `progress` = problems solved this period.
- [x] `GET /consistency/ladders` returns base ladders with the caller's `solved` overlaid per rung.
- [x] All JSON field names match the Flutter models (snake_case) — the app's mock source can be swapped for
      the remote one with no model changes.
