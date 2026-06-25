# Phase 15 — Articles Feed (NEW, partly external)

**Goal:** Back the mobile **articles** feature (`../CPD_HUB/lib/features/articles`). Today the app pulls
Codeforces blog entries **directly** from the public CF API — no backend needed for that path. This phase
adds CPD Hub's own `GET /api/articles` so the platform can serve **first-party** editorials/tutorials and
(optionally) cache external feeds, per `api.md` §10.

**Depends on:** Phase 2 (migrations), Phase 10 (pagination helper — reuse `limit`/`offset`).
**Risk:** Low. Read-only list with pagination + filters. The external CF fetch stays client-side; don't
re-proxy it unless you also add caching.

---

## The contract (`api.md` §10 — camelCase)
```
Article { id, title, author, source, sourceUrl, excerpt, fullContent, publishedAt(ISO8601), tags:[str], rating }
```
Endpoint:
```
GET /api/articles?limit=10&offset=0&source=cpdhub&tag=dp   → [Article]
```
Query params: `limit` (default 10), `offset` (default 0), `source` (`codeforces`|`leetcode`|`cpdhub`),
`tag` (topic filter). All optional.

> Keep the Codeforces direct-fetch in the client as-is. This endpoint is for CPD Hub–authored content and
> any feeds you choose to ingest server-side. The Flutter `ArticleModel.fromJson` reads the CF
> `blogEntry` shape today; when you point it at `/api/articles`, return the flat `Article` shape above.

---

## Checklist
- [ ] 15.1 Domain entity + repository interface (`internal/domain/article.go`).
- [ ] 15.2 Migration `0012_articles` (`articles`, `article_tags`).
- [ ] 15.3 Repository (`articles_repo.go`) with filter + pagination.
- [ ] 15.4 Handler + route + wire repo in `main.go`.
- [ ] 15.5 (Optional) ingest worker that caches external feeds into `articles` on an interval.

## 15.1 Domain
Copy [`article.go`](./article.go) to `internal/domain/article.go`. `Tags []string`, `PublishedAt` as an
ISO-8601 string. The `ArticleFilter` struct carries the query params so the handler stays thin.

## 15.2 Migration
Copy [`0012_articles.up.sql`](./0012_articles.up.sql) / [`.down.sql`](./0012_articles.down.sql).
- `articles(id PK, title, author, source, source_url, excerpt, full_content, published_at, rating)`.
- `article_tags(article_id FK, tag)` — many-to-many; index `(tag)` for tag filtering.

GIN-index nothing fancy yet; a btree on `published_at` (for ordering) and `article_tags(tag)` is enough.

## 15.3 Repository
Copy [`articles_repo.go`](./articles_repo.go). `List(filter)` builds the WHERE clause from the optional
`source` / `tag`, orders by `published_at DESC`, and applies `LIMIT/OFFSET`. Clamp `limit` to `[1,100]`
(default 10) and `offset >= 0` — reuse the Phase-10 pagination helper if it exists; otherwise clamp inline.
Aggregate tags per article with a second query or `array_agg`.

## 15.4 Delivery
Add `GetArticles`, register `protected.GET("/articles", h.GetArticles)`, add
`Article domain.ArticleRepository` to `Repos` + the `Handler` interface, wire in `main.go`. Parse query
params into `domain.ArticleFilter`.

## 15.5 (Optional) ingest worker
If you want server-side caching of external feeds, add a worker mirroring Phase 6's contest cache: fetch on
an interval, upsert into `articles` with `source='codeforces'`. Keep it best-effort — a fetch failure must
never empty the table or 500 the endpoint. Skip entirely if the client keeps fetching CF directly.

---

## Definition of Done
- [ ] `GET /articles` returns first-party articles, newest first, with `limit`/`offset` honored.
- [ ] `?source=cpdhub` and `?tag=dp` filter correctly; combining them ANDs.
- [ ] `limit` is clamped to ≤100; a huge `limit` can't dump the whole table.
- [ ] Empty table returns `[]` (not null, not 500).
- [ ] JSON field names match `api.md` §10 (camelCase).
