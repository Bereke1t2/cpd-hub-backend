# Phase 9 — Activity Feed, Bookmarks & Info

**Goal:** Make the activity feed **real** — generated from user actions (solve/like/contest) instead of
static rows — add the bookmarks endpoints the mobile bookmarks cubit needs, and turn `info` into a small CMS.
Introduce a lightweight domain-event hook so actions in one place (solve) write to several places (activity,
submissions, daily_solves) without tangling the handlers.

**Depends on:** Phases 4 (actions), 5 (submissions).
**Risk:** Low–Medium. Adds write-side fan-out; keep it in usecases.

---

## Checklist
- [ ] 9.1 Emit an activity row when a user solves/likes (event hook).
- [ ] 9.2 Paginate `GET /activity` (newest first) — reuse the Phase-10 pagination helper.
- [ ] 9.3 Bookmarks: `GET /api/bookmarks`, `POST/DELETE /api/bookmarks/:problemId`.
- [ ] 9.4 Info as CMS: keep `GET /info`; add admin create/update (optional, behind a role).
- [ ] 9.5 Humanize activity timestamps for the client (`"2 min ago"`).

## 9.1 Event hook
The mobile home feed shows lines like *"abel solved 'Two Sum' in 3 min"*. Generate these. Add a minimal
emitter the solve/like usecase calls. Copy [`events.go`](./events.go) to
`internal/usecase/activity/events.go`. It exposes `RecordSolve(username, problem)` / `RecordLike(...)` which
insert an `activity` row (and, for solves, the `submissions` + `daily_solves` rows from Phases 4–5). Keeping
the fan-out in one place means the handler still just calls `MarkSolved`; the usecase wires the side effects.

> Don't call this from the repository (the repo shouldn't know about activity). Call it from the
> problems **usecase** after a successful solve, or have the solve usecase depend on an `ActivityRecorder`
> interface so it's testable and decoupled.

## 9.2 Paginate the feed
`GET /api/activity` currently returns everything. Add `?limit=&offset=` (or cursor by `created_at`). Order by
`created_at DESC`. Default limit 20, max 100. The client list is fine with either a bare array (current) or
the paginated envelope — if you add the envelope, do it only on this endpoint and update the client's parser.
Simplest: keep returning a bare array but cap it server-side and accept `limit`/`offset`.

## 9.3 Bookmarks
Table `bookmarks` already exists (Phase 2, migration `0004`). Copy [`bookmarks.go`](./bookmarks.go) for the
domain interface + repo, and add handlers:
```
GET    /api/bookmarks               → [Problem]   (the caller's bookmarked problems, with state)
POST   /api/bookmarks/:problemId    → { success }
DELETE /api/bookmarks/:problemId    → { success }
```
`GET /bookmarks` reuses the Phase-4 problem read with state, filtered to the caller's bookmarked ids. Match
whatever shape the mobile bookmarks cubit expects (it reuses the `Problem` model, so return `[Problem]`).

## 9.4 Info CMS (optional)
`GET /info` stays. If you want editable announcements, add a `role` column to `users` (`user`/`admin`) and
guard `POST/PUT/DELETE /api/info` with an admin check in middleware. Skip if announcements are seeded only.

## 9.5 Humanize timestamps
The `Activity.timestamp` field is a display string (`"2 min ago"`). Compute it from `created_at` at read time
so it's always current, rather than storing a frozen string. Copy the `humanizeSince` helper in `events.go`.
Keep the real `created_at` for ordering; send the humanized string in `timestamp` for the client.

---

## Definition of Done
- [ ] Solving a problem produces an activity row that appears in `GET /activity`.
- [ ] `GET /activity?limit=10` returns at most 10, newest first.
- [ ] Bookmark add/remove works; `GET /bookmarks` returns the caller's bookmarked problems with correct
      per-user `solved`/`isLiked` state.
- [ ] Activity `timestamp` is a fresh humanized string derived from `created_at`.
- [ ] The solve handler stays thin — the activity/submission fan-out lives in the usecase.
