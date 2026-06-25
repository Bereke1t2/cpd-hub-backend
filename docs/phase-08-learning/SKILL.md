# Phase 8 — Structured Learning (NEW feature)

**Goal:** Back the mobile **learning** feature — the topic dependency graph (DAG), tracks, and per-topic
lessons that currently come from a mock data source. Serve topics with their prerequisite edges and linked
problems, the curated tracks, and lesson content, so the skill tree renders from the server.

**Depends on:** Phase 4 (problems exist; solved state powers "what to learn next").
**Risk:** Low–Medium. Mostly read-only content + a graph. The content is curated/seeded, not user-generated.

---

## The contract (must match the Flutter models — snake_case)
From `../CPD_HUB/lib/features/learning/data/models/`:

```
Topic  { id, name, category, summary, difficulty:int,
         prerequisite_ids:[str], problem_ids:[str], reference_urls:[str] }
Track  { id, title, description, topic_ids:[str], icon_name(=school) }
Lesson { topic_id, body, key_ideas:[str] }
```

Endpoints:
```
GET /api/learning/topics            → [Topic]    (the full DAG)
GET /api/learning/tracks            → [Track]
GET /api/learning/lessons/:topicId  → Lesson     (404 if no lesson for the topic)
```

> The client's `learning_path_engine.dart` does the "what to learn next/before" classification **on-device**
> from the topic graph + the user's solved set. So the backend just needs to serve a correct, acyclic graph;
> the recommendation logic stays client-side (Phase 11 can add a server-side recommender if desired).

---

## Checklist
- [ ] 8.1 Domain entities + repository interface (`internal/domain/learning.go`).
- [ ] 8.2 Migration `0009_learning` (`topics`, `topic_prerequisites`, `topic_problems`, `topic_references`,
      `tracks`, `track_topics`, `lessons`).
- [ ] 8.3 Repository (`learning_repo.go`) — assemble each topic's arrays from the edge tables.
- [ ] 8.4 Handlers + routes + wire in `main.go`.
- [ ] 8.5 Seed the topic graph (this is the real work — a curated CP curriculum).
- [ ] 8.6 Validate the graph is acyclic at seed time.

## 8.1 Domain
Copy [`learning.go`](./learning.go) to `internal/domain/learning.go`. snake_case JSON to match the client.

## 8.2 Migration
Copy [`0009_learning.up.sql`](./0009_learning.up.sql) / `.down.sql`. The arrays on `Topic`
(`prerequisite_ids`, `problem_ids`, `reference_urls`) are modeled as **edge tables**, not Postgres arrays, so
they're queryable and FK-checked:
- `topics(id, name, category, summary, difficulty)`
- `topic_prerequisites(topic_id, prerequisite_id)` — the DAG edges
- `topic_problems(topic_id, problem_id)` — links to the `problems` table
- `topic_references(topic_id, url)`
- `tracks(id, title, description, icon_name)` + `track_topics(track_id, topic_id, ord)`
- `lessons(topic_id PK, body, key_ideas)` — `key_ideas` can be a JSON/text array column (small, read-only).

## 8.3 Repository
Copy [`learning_repo.go`](./learning_repo.go). `GetTopics()` loads all topics, then fills each topic's three
arrays from the edge tables (do it in a few batched queries, not N+1 — the template selects all edges once and
groups in Go). `GetLesson(topicId)` returns the lesson or `ErrNotFound`.

## 8.4 Delivery
Add `GetTopics/GetTracks/GetLesson` handlers, register under a `/learning` protected group, add
`Learning domain.LearningRepository` to `Repos` + the `Handler` interface, wire in `main.go`.

## 8.5 Seed — the curriculum
This is the substance of the phase. Author a starter CP topic graph in `cmd/seed` (or a dedicated
`seed_learning.sql`). A reasonable starter set with edges:
```
implementation → (none)
math-basics    → (none)
sorting        → implementation
binary-search  → sorting
two-pointers   → sorting
prefix-sums    → implementation
greedy         → sorting
graphs-bfs-dfs → implementation
dp-intro       → math-basics, greedy
dp-knapsack    → dp-intro
shortest-paths → graphs-bfs-dfs
segment-tree   → prefix-sums, binary-search
```
Each topic links a few `problem_ids` (must exist in `problems`) and 1–2 reference URLs. Group topics into 2–3
tracks (e.g. "Beginner Foundations", "Intermediate", "Graphs & Trees").

## 8.6 Acyclicity check
A prerequisite cycle would hang the client's path engine. Add a tiny validation at seed time (or a startup
check): topological sort the edges; fail loudly if a cycle is found. The template includes `assertAcyclic` you
can call from `cmd/seed`.

---

## Definition of Done
- [ ] `GET /learning/topics` returns the full graph; each topic has correct `prerequisite_ids`, `problem_ids`,
      `reference_urls`.
- [ ] `GET /learning/tracks` returns curated tracks with ordered `topic_ids`.
- [ ] `GET /learning/lessons/:topicId` returns a lesson, or 404 for a topic with no lesson.
- [ ] The seeded graph is acyclic (verified at seed time).
- [ ] JSON matches the Flutter models — the app's mock learning source can be swapped for remote unchanged.
