# Phase 5 — Profiles & Analytics

**Goal:** Make the profile analytics endpoints return **real** data from the tables Phase 2 added
(`submissions`, `rating_history`, `daily_solves`, `attendance`) instead of the static samples in the current
handler. Compute profile aggregates (`solvedProblems`, `attendedContestsCount`, etc.) and serve the heatmap
from `daily_solves`.

**Depends on:** Phases 2 (tables), 4 (solves populate `daily_solves`).
**Risk:** Medium. Read-heavy; no destructive changes. Can run in parallel with Phase 6.

---

## Checklist
- [x] 5.1 Implement the four analytics repo methods against the new tables.
- [x] 5.2 Compute profile aggregates in `GetProfile`.
- [x] 5.3 Record submissions when a solve happens (link to Phase 4 / Phase 9 events).
- [x] 5.4 Seed analytics rows so dev profiles render.
- [x] 5.5 Remove the static fallbacks from the handlers.

---

## 5.1 Analytics methods
Copy [`profile_repo.go`](./profile_repo.go) to `internal/infrastructure/databases/profile_repo.go` (or merge
into the existing one). It implements:

- `GetProfileHeatmap(username)` → `SELECT to_char(day,'YYYY-MM-DD'), count FROM daily_solves WHERE username=$1
  ORDER BY day` → `[]HeatmapEntry`. This is the calendar the profile page draws.
- `GetProfileRatingHistory(username)` → from `rating_history`.
- `GetProfileAttendance(username)` → from `attendance`.
- `GetProfileSubmissions(username)` → from `submissions` (newest first, capped — paginate in Phase 10).

Each returns an **empty slice, not an error**, when there's no data — the client renders an empty heatmap
fine, but a 500 breaks the page. (The current handler returns 500 on repo error; with real empty tables that
would fire constantly.)

## 5.2 Profile aggregates
`GetProfile` should fill the computed fields the `UserProfile` entity already has:
```sql
SELECT
  u.username, u.full_name, COALESCE(p.bio,''), COALESCE(p.avatar_url,''), COALESCE(p.rating,0),
  (SELECT count(*) FROM user_problems WHERE username=u.username AND solved)        AS solved_problems,
  (SELECT count(*) FROM attendance   WHERE username=u.username AND status='Present') AS attended
FROM users u LEFT JOIN profiles p ON p.username = u.username
WHERE u.username = $1
```
Map to `SolvedProblems`, `AttendedContestsCount`, etc. `GlobalRank` can be a windowed rank over
`profiles.rating` (or deferred). Don't invent values the UI doesn't need yet.

## 5.3 Recording submissions
A submission row should be created whenever a user solves (or attempts) a problem. The cleanest hook is the
solve action from Phase 4: when `MarkSolved` succeeds, also insert a `submissions` row with
`status='Accepted'`. In Phase 9 this becomes a proper domain event that also writes the activity feed. For
now, a direct insert in the solve usecase is fine — keep it in the usecase, not the handler.

## 5.4 Seed
Add to `cmd/seed`: a handful of `daily_solves` rows across the last ~60 days for the demo user, a few
`rating_history` points, and some `submissions` so the profile page isn't blank in dev. Use relative dates
computed in Go (e.g. `time.Now().AddDate(0,0,-i)`), not hardcoded 2026 strings.

## 5.5 Drop the static fallbacks
In `handler.go`, the `profileHeatmap`/`profileRatingHistory`/`profileAttendance`/`profileSubmissions` and
`getProfile`/`listUsers` handlers each have a hardcoded sample branch. Remove them; rely on the repo. Map repo
errors through `respondError` (Phase 0). The success responses stay the same raw arrays/objects the client
parses.

---

## Definition of Done
- [x] Profile heatmap reflects the caller's actual `daily_solves` (verify by solving a problem then GETting
      the heatmap).
- [x] Rating history, attendance, submissions come from their tables; empty data returns `[]`, not 500.
- [x] `GET /users/profile/:username` includes a real `solvedProblems` count.
- [x] No hardcoded `2026-02-01` sample rows remain in `handler.go`.
- [x] Dev seed makes the demo profile render with a populated heatmap.
