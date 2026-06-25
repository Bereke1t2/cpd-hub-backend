# Phase 11 — Testing & CI/CD

**Goal:** Lock in correctness with a real test pyramid — fast usecase unit tests, handler tests with
`httptest` + mock repos, and repository tests against a throwaway Postgres — then run it all in GitHub Actions
on every push. Ideally you wrote tests inside each prior phase; this phase fills gaps and adds the pipeline.

**Depends on:** Phases 1–10.
**Risk:** Low. Tests don't change behavior; they prevent regressions.

---

## The pyramid
```
        /\        repo tests (dockertest Postgres)  — few, slow, high-value (SQL correctness)
       /  \       handler tests (httptest + mocks)  — medium
      /____\      usecase tests (pure, table-driven) — many, fast
```

## Checklist
- [ ] 11.1 Usecase unit tests (streak math, auth handle derivation, daily pick) — pure, no DB.
- [ ] 11.2 Mock repositories (hand-written or `testify/mock`) for handler tests.
- [ ] 11.3 Handler tests with `httptest` covering happy path + 400/401/404.
- [ ] 11.4 Repository tests against a real Postgres via `dockertest` (or a CI service container).
- [ ] 11.5 GitHub Actions: build, vet, test (with a Postgres service), lint.
- [ ] 11.6 Coverage gate (start at ~60% on `internal/usecase`, raise over time).

## 11.1 Usecase tests
The streak engine (Phase 7) and auth handle derivation (Phase 3) are pure functions — ideal for table-driven
tests. Copy [`usecase_test.go`](./usecase_test.go) to `internal/usecase/consistency/consistency_test.go` and
adapt. Cover: empty history, single day, a freeze bridging one gap, two gaps breaking the streak, solving
today vs not-yet-today.

## 11.2 Mocks
Hand-written mocks are fine and dependency-free. A mock repo records calls and returns canned data:
```go
type mockProblemRepo struct {
	liked string
	err   error
}
func (m *mockProblemRepo) Like(username, id string) error { m.liked = id; return m.err }
// ...implement the rest of the interface
```
Put them in a `_test.go` so they don't ship.

## 11.3 Handler tests
Copy [`handler_test.go`](./handler_test.go) to `internal/delivery/httpdelivery/handler_test.go`. It builds a
gin engine with mock repos, fires requests via `httptest.NewRecorder`, and asserts status + JSON. Cover the
bug we fixed in Phase 1 (like returns **200**, not 500) so it can never regress, plus 404 for unknown ids and
401 without a token.

## 11.4 Repository tests
SQL correctness (the per-user toggle transaction, the daily pick, the heatmap query) can only be tested
against a real Postgres. Use `ory/dockertest` to spin one up, run migrations, exercise the repo, assert. These
are slower — tag them `//go:build integration` and run them in CI + on demand, not on every `go test ./...`.

## 11.5 GitHub Actions
Copy [`ci.yml`](./ci.yml) to `.github/workflows/ci.yml`. It:
- checks out, sets up Go, caches modules;
- runs `go vet`, `gofmt -l` (fails if anything unformatted), `go build ./...`;
- runs `go test ./...` with a Postgres **service container** so integration tests have a DB;
- runs `golangci-lint`.

## 11.6 Coverage
Add `go test ./... -coverprofile=cover.out` and print `go tool cover -func`. Gate the merge on a floor for
`internal/usecase` (where the logic lives). Don't chase 100% — chase the branches that matter (streak edges,
auth failures, toggle transitions).

---

## Definition of Done
- [ ] `go test ./...` passes locally and is green in CI.
- [ ] A test asserts like/solve return 200 (guards the Phase-1 fix).
- [ ] Streak math has table-driven tests covering freeze + gap cases.
- [ ] At least one repo integration test runs the real toggle transaction against Postgres.
- [ ] CI runs build + vet + fmt + test + lint on every push/PR.
