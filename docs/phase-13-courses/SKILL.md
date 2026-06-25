# Phase 13 — Courses (NEW feature)

**Goal:** Build the backend for the mobile **courses** feature (`../CPD_HUB/lib/features/courses`) which
today runs on mock data. Serve structured learning modules (video / article / pdf lessons) and persist
per-user lesson completion so progress survives reinstalls and syncs across devices.

**Depends on:** Phase 3 (auth identity / `loadUser`), Phase 4 (per-user state pattern).
**Risk:** Low–Medium. Net-new domain, pure read + one write (`complete`). Follows the Phase-0 feature recipe.

---

## The contract (must match `api.md` §8 and the Flutter models — camelCase)
From `../CPD_HUB/lib/features/courses/data/models/course_model.dart`:

```
Course { id, title, summary, level, modules:[Module] }
Module { id, title, lessons:[Lesson] }
Lesson { id, title, kind("video"|"article"|"pdf"), contentUrl, inlineText?, durationSec?, completed }
```

Endpoints:
```
GET  /api/courses                                          → [Course]   (progress merged for caller)
GET  /api/courses/:id                                      → Course     (full detail + progress)
POST /api/courses/:courseId/lessons/:lessonId/complete     → { lessonId, completed:true }
```

> `completed` is **per-user** — it is never a column on `lessons`. It is overlaid from
> `user_lesson_progress` keyed by the authenticated username, exactly like `solved` in Phase 4.

---

## Checklist
- [ ] 13.1 Domain entities + repository interface (`internal/domain/course.go`).
- [ ] 13.2 Migration `0010_courses` (`courses`, `course_modules`, `course_lessons`, `user_lesson_progress`).
- [ ] 13.3 Repository (`courses_repo.go`) — load tree, overlay caller's completion.
- [ ] 13.4 Usecase only if needed (pure pass-through here — skip unless you add ordering/recommendations).
- [ ] 13.5 Handlers + routes + wire repo in `main.go`.
- [ ] 13.6 Seed one or two courses in `cmd/seed` so `GET /courses` returns content.

## 13.1 Domain
Copy [`course.go`](./course.go) to `internal/domain/course.go`. JSON tags are camelCase to match the client.
`Lesson.Completed` is computed per request, not stored on the entity in the DB. `Lesson.InlineText` and
`Lesson.DurationSec` are `omitempty` because they only apply to `article` / `video` kinds respectively.

## 13.2 Migration
Copy [`0010_courses.up.sql`](./0010_courses.up.sql) / [`.down.sql`](./0010_courses.down.sql) to `migrations/`.
Tables:
- `courses(id PK, title, summary, level)`.
- `course_modules(id PK, course_id FK, title, ord)` — `ord` preserves module order.
- `course_lessons(id PK, module_id FK, title, kind, content_url, inline_text, duration_sec, ord)`.
- `user_lesson_progress(username, lesson_id, completed_at)` — PK `(username, lesson_id)`; the per-user overlay.

`ord` columns matter: the client renders modules and lessons in insertion order, so always `ORDER BY ord`.

## 13.3 Repository
Copy [`courses_repo.go`](./courses_repo.go). Two reads:
- `List(username)` — load all courses → modules → lessons in three ordered queries (avoid N+1: fetch all
  rows once, then assemble the tree in Go), then `LEFT JOIN`/overlay `user_lesson_progress` for `completed`.
- `Get(username, id)` — same, scoped to one course id; return `domain.ErrNotFound` if the id is unknown.

`CompleteLesson(username, courseId, lessonId)` is an idempotent upsert into `user_lesson_progress`
(`ON CONFLICT DO NOTHING`). Validate the lesson belongs to the course first → `404` if not.

## 13.4 Usecase
Skip — this is a pass-through. Add a usecase only if you later compute a "next lesson" / percent-complete
field. Keep the handler thin and call the repo directly, matching `problemsList` today.

## 13.5–13.6 Delivery + seed
Add `GetCourses / GetCourse / CompleteLesson` handlers, register routes under a `/courses` protected group,
add `Course domain.CourseRepository` to `Repos` + the `Handler` interface, wire in `main.go`. Seed at least
one full course (module + a video lesson + an article lesson) in `cmd/seed`.

---

## Definition of Done
- [ ] `GET /courses` returns courses with nested modules/lessons in `ord` order, `completed` reflecting the
      caller's real progress.
- [ ] `GET /courses/:id` returns one course; unknown id → `404 not_found` (not 500).
- [ ] `POST /courses/:courseId/lessons/:lessonId/complete` is idempotent and flips `completed` for that user
      only; a second call returns the same `{lessonId, completed:true}` without error.
- [ ] A lesson id that doesn't belong to the course → `404`.
- [ ] All JSON field names match the Flutter `CourseModel` (camelCase); the app's mock source swaps for the
      remote one with no model changes.
