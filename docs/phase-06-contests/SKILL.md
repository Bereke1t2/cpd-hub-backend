# Phase 6 — Contests Integration

**Goal:** Harden the contests path: cache the external Kontests/Codeforces calls, never let an upstream
failure 500 the screen, serve the fields the mobile **countdown** UI needs (`startTime`, `isPast`,
`isParticipating`), and add a background refresh so the list is fast and fresh.

**Depends on:** Phases 2, 3.
**Risk:** Medium. External APIs are flaky and rate-limited — caching + graceful degradation is the whole job.

---

## Checklist
- [x] 6.1 Wrap the Kontests client with an in-memory TTL cache.
- [x] 6.2 Graceful degradation: on upstream error, serve last-good cache or the DB list, never 500.
- [x] 6.3 Ensure `startTime` (RFC3339), `duration`, `isPast` are always set for the countdown.
- [x] 6.4 Per-user participation: `isParticipating` from a `contest_participants` join.
- [x] 6.5 Cache leaderboard responses (Codeforces standings) with a short TTL.
- [x] 6.6 Optional: background refresh worker that warms the cache every N minutes.

---

## 6.1 Cache the client
The handler currently does `external.NewKontestsClient()` **per request** and fetches live every time. Wrap
it. Copy [`cache.go`](./cache.go) to `internal/infrastructure/external/cache.go` (a tiny generic TTL cache)
and [`cached_contests.go`](./cached_contests.go) to `internal/infrastructure/external/cached_contests.go`.
Build the client + cache **once** in `main.go` and inject it into the contests usecase, rather than
constructing it in the handler.

## 6.2 Degrade gracefully
The contests usecase should:
1. Try the cache → return if fresh.
2. On miss, fetch upstream. On success, cache + return.
3. On upstream error, return the **stale** cache if present, else the DB-backed list, else `[]` with a logged
   warning. Mobile shows whatever it has; it must not see a 500 because Codeforces had a blip.

See `cached_contests.go`'s `List()` for this fallback ladder.

## 6.3 Countdown fields
The mobile countdown (mobile Phase 16) needs a real `startTime`. The `Contest` entity already has
`StartTime time.Time` serialized as `startTime`. Ensure the mapping from Kontests/CF fills it (parse their
ISO start string into `time.Time`), compute `IsPast = startTime.Before(now)`, and set `Duration` from their
length field. Sort upcoming first, past last. Don't send a zero time — the client can't count down to it.

## 6.4 Participation
Add a small join so a user can mark themselves participating:
```sql
CREATE TABLE contest_participants (
    username   TEXT REFERENCES users(username) ON DELETE CASCADE,
    contest_id TEXT NOT NULL,
    PRIMARY KEY (username, contest_id)
);
```
(Ship as `migrations/0007_contest_participants.up.sql`.) Add `POST/DELETE /api/contests/:id/participate`
and fill `isParticipating` per caller in the list mapping. Optional but matches the entity field already
sent to the client.

## 6.5 Leaderboard cache
`contestLeaderboard` hits the Codeforces standings API live. Cache by
`(contestID, from, count, showUnofficial)` for ~60s. Standings change slowly during a contest and not at all
after; a short TTL kills repeated upstream calls when many users open the same leaderboard.

## 6.6 Refresh worker (optional)
Copy [`worker.go`](./worker.go) to `internal/infrastructure/external/worker.go`: a goroutine started from
`main.go` that calls `List()` every N minutes to keep the cache warm, so user requests are always cache hits.
Stop it on shutdown via context (the Phase-1 graceful shutdown passes a cancelable ctx). Keep N modest
(5–10 min) to respect upstream rate limits.

---

## Definition of Done
- [x] `GET /contests` is a cache hit on the second call within the TTL (verify via timing/logs).
- [x] Simulated upstream failure (point the client at a bad URL) still returns 200 with stale/DB data.
- [x] Every contest in the response has a non-zero `startTime` and a correct `isPast`.
- [x] Leaderboard requests for the same contest within the TTL don't re-hit Codeforces.
- [x] (If done) `isParticipating` reflects the caller's participation; participate/unparticipate works.
