# Phase 14 — Smart Practice: Spaced Repetition + Upsolving (NEW feature)

**Goal:** Build the backend for the mobile **practice** feature (`../CPD_HUB/lib/features/practice`). Two
sub-features, both per-user and today mock-only:
1. **Review queue** — SM-2 spaced-repetition cards for problems the user wants to retain.
2. **Upsolves** — problems flagged from past contests to solve later.

**Depends on:** Phase 3 (auth identity / `loadUser`), Phase 4 (per-user state pattern).
**Risk:** Medium. The SM-2 scheduling math is the only real logic — it lives in a usecase and is unit-tested
here (don't wait for Phase 11).

---

## The contract (must match `api.md` §9 and the Flutter models — snake_case)
From `../CPD_HUB/lib/features/practice/data/models/`:

```
ReviewItem  { problem_id, due_date(ISO8601), interval(days), ease(>=1.3, start 2.5), repetitions }
UpsolveItem { contest_id, contest_title, problem_id, problem_title, resolved }
```

Endpoints:
```
GET    /api/practice/review-queue                  → [ReviewItem]   (caller's cards due/all)
POST   /api/practice/review-queue   body ReviewItem → ReviewItem    (201, first add)
PUT    /api/practice/review-queue/:problemId        → ReviewItem     (after a recall grade)
DELETE /api/practice/review-queue/:problemId        → 204
GET    /api/practice/upsolves                       → [UpsolveItem]
POST   /api/practice/upsolves       body UpsolveItem → UpsolveItem   (201)
PUT    /api/practice/upsolves/:problemId            → UpsolveItem    (toggle resolved)
```

> Every row is scoped to the authenticated username. `:problemId` identifies the card/upsolve **within**
> that user's set — never globally.

---

## Checklist
- [ ] 14.1 Domain entities + repository interface (`internal/domain/practice.go`).
- [ ] 14.2 Migration `0011_practice` (`review_items`, `upsolve_items`).
- [ ] 14.3 SM-2 usecase (`internal/usecase/practice/sm2.go`) + unit test.
- [ ] 14.4 Repository (`practice_repo.go`).
- [ ] 14.5 Handlers + routes + wire repo in `main.go`.

## 14.1 Domain
Copy [`practice.go`](./practice.go) to `internal/domain/practice.go`. snake_case JSON tags. `DueDate` is an
ISO-8601 string (the client sends/reads `toIso8601String`); store as `TIMESTAMPTZ` and format on the way out
with `time.RFC3339`. `Ease` is a `float64` (SM-2 ease factor).

## 14.2 Migration
Copy [`0011_practice.up.sql`](./0011_practice.up.sql) / [`.down.sql`](./0011_practice.down.sql).
- `review_items(username, problem_id, due_date, interval, ease, repetitions)` — PK `(username, problem_id)`.
- `upsolve_items(username, problem_id, contest_id, contest_title, problem_title, resolved)` — PK
  `(username, problem_id)`.

Index `review_items(username, due_date)` so "due today" queries stay cheap.

## 14.3 SM-2 usecase (the only real logic)
Copy [`sm2.go`](./sm2.go) to `internal/usecase/practice/sm2.go`. The standard SuperMemo-2 update, given a
recall `quality` 0–5:

```
if quality < 3:               // failed recall
    repetitions = 0
    interval    = 1
else:
    repetitions += 1
    if repetitions == 1: interval = 1
    elif repetitions == 2: interval = 6
    else: interval = round(interval * ease)
ease = max(1.3, ease + (0.1 - (5-quality)*(0.08 + (5-quality)*0.02)))
due_date = today + interval days
```

The server is the source of truth for scheduling: the client sends the `quality` grade (or the already-
computed item — accept both, but **recompute server-side** from `quality` when present so a tampered client
can't game the schedule). Copy [`sm2_test.go`](./sm2_test.go) and keep it green.

## 14.4 Repository
Copy [`practice_repo.go`](./practice_repo.go). Plain CRUD upserts keyed by `(username, problem_id)`:
- `ListReviewQueue(username)` — all of the user's cards, `ORDER BY due_date`.
- `AddReview / UpdateReview` — `INSERT ... ON CONFLICT (username, problem_id) DO UPDATE`.
- `DeleteReview(username, problemID)` — `RowsAffected()==0 → ErrNotFound`.
- Upsolve equivalents; `UpdateUpsolve` only flips `resolved`.

## 14.5 Delivery
Add the seven handlers, register under a `/practice` protected group, add
`Practice domain.PracticeRepository` to `Repos` + the `Handler` interface, wire in `main.go`. On `PUT
/review-queue/:problemId`, if the body carries a `quality` field, run it through the SM-2 usecase before
saving; otherwise persist the client-supplied item as-is.

---

## Definition of Done
- [ ] `POST /practice/review-queue` creates a card (201) with `ease=2.5, interval=1, repetitions=0` defaults
      when omitted.
- [ ] `PUT /practice/review-queue/:problemId` with a passing grade pushes `due_date` out and grows `interval`
      per SM-2; a failing grade resets `interval` to 1 and `repetitions` to 0.
- [ ] `DELETE` removes only the caller's card; unknown id → `404`.
- [ ] Upsolve add/list/toggle work and are scoped to the caller.
- [ ] `go test ./internal/usecase/practice/...` passes (SM-2 unit test).
- [ ] All JSON field names match the Flutter models (snake_case).
