# Feature Scaffold — the 5-file recipe

Copy these skeletons when adding any new feature. Replace `Widget` with your aggregate name
(e.g. `Streak`, `Topic`, `Bookmark`). Order matters: build keeps compiling at every step.

Module path is `github.com/bereket/cpd-hub-backend`.

---

## 1. Domain — `internal/domain/widget.go`

```go
package domain

type Widget struct {
	ID        string `json:"id"`
	OwnerName string `json:"-"` // the username it belongs to; never leak in JSON unless intended
	Name      string `json:"name"`
}

// WidgetRepository is implemented in infrastructure/databases.
type WidgetRepository interface {
	GetForUser(username string) (*Widget, error)
	SaveForUser(username string, w *Widget) error
	List() ([]*Widget, error)
}
```

## 2. Migration — `migrations/000N_widget.up.sql` (Phase 2 onward)

```sql
CREATE TABLE IF NOT EXISTS widgets (
    id         TEXT PRIMARY KEY,
    username   TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_widgets_username ON widgets(username);
```

…and `migrations/000N_widget.down.sql`:

```sql
DROP TABLE IF EXISTS widgets;
```

## 3. Repository — `internal/infrastructure/databases/widget_repo.go`

```go
package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type WidgetRepositoryDB struct{ client *postgres.Client }

func NewWidgetRepositoryDB(c *postgres.Client) *WidgetRepositoryDB {
	return &WidgetRepositoryDB{client: c}
}

func (r *WidgetRepositoryDB) GetForUser(username string) (*domain.Widget, error) {
	row := r.client.Pool.QueryRow(context.Background(),
		`SELECT id, username, name FROM widgets WHERE username=$1`, username)
	var w domain.Widget
	if err := row.Scan(&w.ID, &w.OwnerName, &w.Name); err != nil {
		return nil, domain.ErrNotFound("widget not found").Wrap(err)
	}
	return &w, nil
}

func (r *WidgetRepositoryDB) SaveForUser(username string, w *domain.Widget) error {
	_, err := r.client.Pool.Exec(context.Background(),
		`INSERT INTO widgets (id, username, name) VALUES ($1,$2,$3)
		 ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, updated_at=now()`,
		w.ID, username, w.Name)
	if err != nil {
		return domain.ErrInternal("could not save widget").Wrap(err)
	}
	return nil
}

func (r *WidgetRepositoryDB) List() ([]*domain.Widget, error) { /* ... */ return nil, nil }
```

## 4. Usecase — `internal/usecase/widget/widget_usecase.go`

Only needed when there's logic beyond a pass-through. Keep it gin-free and sql-free.

```go
package widget

import "github.com/bereket/cpd-hub-backend/internal/domain"

type UseCase struct{ repo domain.WidgetRepository }

func New(repo domain.WidgetRepository) *UseCase { return &UseCase{repo: repo} }

func (uc *UseCase) GetOrDefault(username string) (*domain.Widget, error) {
	w, err := uc.repo.GetForUser(username)
	if err != nil {
		// first-time user gets a sensible default rather than a 404
		return &domain.Widget{ID: username + "-default", OwnerName: username, Name: "New"}, nil
	}
	return w, nil
}
```

## 5. Delivery — handler methods + routes

Add handler methods (in `handler.go` or a `widget_handler.go` in the same package):

```go
func (h *handlerImpl) GetWidget(c *gin.Context) {
	username := currentUsername(c) // from loadUser middleware (Phase 3)
	uc := widget.New(h.repos.Widget)
	w, err := uc.GetOrDefault(username)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, w)
}
```

Register in `routes.go` under `protected`:

```go
protected.GET("/widgets/me", h.GetWidget)
```

Wire the repo in `cmd/server/main.go`:

```go
repos := httpdelivery.Repos{
	// ...existing...
	Widget: databases.NewWidgetRepositoryDB(client),
}
```

…and add `Widget domain.WidgetRepository` to the `Repos` struct + the `Handler` interface method
`GetWidget(*gin.Context)`.

---

### Checklist for any new feature
- [ ] Domain entity + repository interface (stdlib-only imports).
- [ ] Up + down migration; `go run ./cmd/migrate up` applies cleanly.
- [ ] Repository implements the interface; errors are `*domain.AppError`.
- [ ] Usecase holds the logic (or skip if pure pass-through).
- [ ] Handler is thin; routes registered; repo wired in `main.go`.
- [ ] Path + JSON field names match `../CPD_HUB/lib/core/url_constants.dart` and the model `fromJson`.
- [ ] `go build ./...` and `go vet ./...` clean.
