//go:build ignore
// Template for Phase 9 — copy to: internal/domain/bookmark.go + a repo in databases.
//
// Backs the mobile bookmarks cubit. GET returns the caller's bookmarked problems
// (reusing the Phase-4 problem-with-state read), POST/DELETE toggle membership.

// ---- internal/domain/bookmark.go ----
package domain

type BookmarkRepository interface {
	Add(username, problemID string) error
	Remove(username, problemID string) error
	ListProblemIDs(username string) ([]string, error)
}

/* ---- internal/infrastructure/databases/bookmark_repo.go ----

package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type BookmarkRepositoryDB struct{ client *postgres.Client }

func NewBookmarkRepositoryDB(c *postgres.Client) *BookmarkRepositoryDB {
	return &BookmarkRepositoryDB{client: c}
}

func (r *BookmarkRepositoryDB) Add(username, problemID string) error {
	_, err := r.client.Pool.Exec(context.Background(),
		`INSERT INTO bookmarks (username, problem_id) VALUES ($1,$2)
		 ON CONFLICT DO NOTHING`, username, problemID)
	if err != nil {
		return domain.ErrInternal("could not add bookmark").Wrap(err)
	}
	return nil
}

func (r *BookmarkRepositoryDB) Remove(username, problemID string) error {
	_, err := r.client.Pool.Exec(context.Background(),
		`DELETE FROM bookmarks WHERE username=$1 AND problem_id=$2`, username, problemID)
	if err != nil {
		return domain.ErrInternal("could not remove bookmark").Wrap(err)
	}
	return nil
}

func (r *BookmarkRepositoryDB) ListProblemIDs(username string) ([]string, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT problem_id FROM bookmarks WHERE username=$1 ORDER BY created_at DESC`, username)
	if err != nil {
		return nil, domain.ErrInternal("").Wrap(err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out, nil
}

---- handlers (httpdelivery) ----

func (h *handlerImpl) ListBookmarks(c *gin.Context) {
	username := currentUsername(c)
	ids, err := h.repos.Bookmark.ListProblemIDs(username)
	if err != nil { respondError(c, err); return }
	out := make([]gin.H, 0, len(ids))
	for _, id := range ids {
		if p, err := h.repos.Problem.GetByIDForUser(username, id); err == nil {
			out = append(out, apiProblem(p))
		}
	}
	respondOK(c, out)
}

func (h *handlerImpl) AddBookmark(c *gin.Context) {
	if err := h.repos.Bookmark.Add(currentUsername(c), c.Param("problemId")); err != nil {
		respondError(c, err); return
	}
	respondSuccess(c)
}

func (h *handlerImpl) RemoveBookmark(c *gin.Context) {
	if err := h.repos.Bookmark.Remove(currentUsername(c), c.Param("problemId")); err != nil {
		respondError(c, err); return
	}
	respondSuccess(c)
}

---- routes ----
	bm := protected.Group("/bookmarks")
	bm.GET("", h.ListBookmarks)
	bm.POST("/:problemId", h.AddBookmark)
	bm.DELETE("/:problemId", h.RemoveBookmark)
*/
