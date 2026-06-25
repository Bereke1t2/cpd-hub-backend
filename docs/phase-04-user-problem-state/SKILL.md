# Phase 4 — Per-User Problem State

**Goal:** Make like / dislike / solve **per-user** instead of global. Today these flip a column on the shared
`problems` row, so every user sees the same `isLiked`/`solved`. Move per-user state into the `user_problems`
join (Phase 2), compute `isLiked`/`isDisliked`/`solved` for the **calling** user, keep `numberOfLikes` etc.
as aggregate counts, and make the daily problem date-based. Also record a `daily_solves` row on solve so the
heatmap (Phase 5) has data.

**Depends on:** Phases 2 (schema), 3 (loadUser → `currentUsername`).
**Risk:** High. Every problem read/write changes. The client JSON shape must stay identical
(`isLiked`, `isDisliked`, `solved`, `numberOfLikes`, `numberOfDislikes`, `numberOfSolvedPeople`).

---

## Checklist
- [x] 4.1 Extend `ProblemRepository` with user-scoped methods.
- [x] 4.2 Implement them against `user_problems` (copy `user_problem_repo.go`).
- [x] 4.3 Like/dislike/solve usecase: upsert join row; keep counts consistent in a transaction.
- [x] 4.4 Reads (`List`, `GetById`, `GetDaily`) join the current user's state + counts.
- [x] 4.5 Handlers pass `currentUsername(c)` into every problem call.
- [x] 4.6 `apiProblem` fills `numberOfSolvedPeople` from a real count.
- [x] 4.7 Date-based daily problem (deterministic per day).
- [x] 4.8 On solve, increment `daily_solves(username, today)`.

---

## 4.1 Repository surface
The interface methods become user-aware. Change `domain.ProblemRepository`:
```go
type ProblemRepository interface {
	ListForUser(username string) ([]*Problem, error)
	GetByIDForUser(username, id string) (*Problem, error)
	GetDailyForUser(username string) (*Problem, error)

	Like(username, id string) error
	Dislike(username, id string) error
	MarkSolved(username, id string) error
	UnmarkSolved(username, id string) error

	CountSolvers(id string) (int, error)
}
```
`Problem` gains no new JSON fields — `IsLiked/IsDisliked/Solved` already exist; they're now filled per user.
Add an unexported `SolverCount int` (json `-`) if you prefer carrying it through, or fill it in the handler.

## 4.2 Repository implementation
Copy [`user_problem_repo.go`](./user_problem_repo.go) to
`internal/infrastructure/databases/user_problem_repo.go`, or fold these queries into the existing
`problems_repo.go`. Reads `LEFT JOIN user_problems` on `(username, id)` so a user who never touched a problem
gets `false`s. The like/dislike toggles move from the global `problems` row to the join row.

**Counts decision:** two options —
1. **Denormalized** (keep `problems.likes`/`dislikes`): update both the join row and the counter in one
   transaction. Fast reads, must keep in sync.
2. **Computed**: `SELECT count(*) FROM user_problems WHERE problem_id=$1 AND liked`. Always correct, slightly
   heavier. With the partial index from Phase 2 this is fine at this scale.

The template uses **option 1** (denormalized counters in a transaction) because the client shows counts on
every list row and you don't want N count queries per list. Pick one and be consistent.

## 4.3 Toggle logic (transaction)
Like is a toggle and is mutually exclusive with dislike. In one `BEGIN…COMMIT`:
- upsert `user_problems` row, flip `liked`, clear `disliked`;
- adjust `problems.likes` (+1/−1) and `problems.dislikes` (−1 if it was disliked).

See the `Like`/`Dislike` methods in the template — they use `INSERT … ON CONFLICT … DO UPDATE` and a single
`UPDATE problems` guarded by the previous state. Wrap both statements in `pool.Begin(ctx)` / `tx.Commit`.

## 4.4–4.6 Reads + solver count
`ListForUser`/`GetByIDForUser` select problem columns plus
`COALESCE(up.liked,false), COALESCE(up.disliked,false), COALESCE(up.solved,false)`. For
`numberOfSolvedPeople`, either select `(SELECT count(*) FROM user_problems WHERE problem_id=p.id AND solved)`
inline, or call `CountSolvers`. Update `apiProblem` in `handler.go` to read the real value instead of the
hardcoded `0`.

## 4.7 Date-based daily
Replace `GetDaily → list[0]`. Pick deterministically by day so all users see the same daily and it's stable
within the day:
```sql
SELECT * FROM problems
ORDER BY md5(id || $1)        -- $1 = today's date 'YYYY-MM-DD'
LIMIT 1
```
This rotates daily without a cron. (Alternative: a `daily_problems(day, problem_id)` table you fill nightly —
overkill for now.) Then layer the calling user's `solved`/`liked` state on top.

## 4.8 Record solves for the heatmap
When `MarkSolved` flips `solved` from false→true, also:
```sql
INSERT INTO daily_solves (username, day, count) VALUES ($1, CURRENT_DATE, 1)
ON CONFLICT (username, day) DO UPDATE SET count = daily_solves.count + 1;
```
Only on the false→true transition (don't double-count re-solves). `UnmarkSolved` may decrement, or leave the
historical count — decide and document. This is what makes the Phase 5 heatmap real.

---

## Definition of Done
- [x] Two different users see independent `isLiked`/`solved` for the same problem.
- [x] `numberOfLikes`/`numberOfDislikes` reflect aggregate counts and stay correct after toggles.
- [x] `numberOfSolvedPeople` is a real count, not `0`.
- [x] Liking then liking again is a clean toggle; liking clears any dislike (and vice-versa).
- [x] `GET /problems/daily` returns the same problem all day, different across days, with the caller's state.
- [x] Solving a new problem creates/increments today's `daily_solves` row.
- [x] Client JSON shape unchanged — the Flutter app needs no edits.
