//go:build ignore

// Template for Phase 11 — copy to: internal/delivery/httpdelivery/handler_test.go
//
// Handler tests with httptest + a mock repo. Guards the Phase-1 fix: a successful
// like must return 200, not 500.
//
package httpdelivery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type mockProblemRepo struct {
	likeErr error
	liked   string
}

func (m *mockProblemRepo) ListForUser(string) ([]*domain.Problem, error) { return nil, nil }
func (m *mockProblemRepo) GetByIDForUser(string, string) (*domain.Problem, error) {
	return &domain.Problem{}, nil
}
func (m *mockProblemRepo) GetDailyForUser(string) (*domain.Problem, error) {
	return &domain.Problem{}, nil
}
func (m *mockProblemRepo) Like(_, id string) error           { m.liked = id; return m.likeErr }
func (m *mockProblemRepo) Dislike(string, string) error      { return nil }
func (m *mockProblemRepo) MarkSolved(string, string) error   { return nil }
func (m *mockProblemRepo) UnmarkSolved(string, string) error { return nil }
func (m *mockProblemRepo) CountSolvers(string) (int, error)  { return 0, nil }

func newTestHandler(repo domain.ProblemRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	g := gin.New()
	h := &handlerImpl{repos: Repos{Problem: repo}, router: g}
	// inject a username as loadUser would
	g.Use(func(c *gin.Context) { c.Set("username", "alice"); c.Next() })
	g.POST("/problems/:id/like", h.LikeProblem)
	return g
}

func TestLikeProblem_Returns200(t *testing.T) {
	repo := &mockProblemRepo{}
	r := newTestHandler(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/problems/p1/like", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (Phase-1 bug guard)", w.Code)
	}
	if repo.liked != "p1" {
		t.Errorf("repo.Like called with %q, want p1", repo.liked)
	}
}

func TestLikeProblem_NotFound(t *testing.T) {
	repo := &mockProblemRepo{likeErr: domain.ErrNotFound("problem not found")}
	r := newTestHandler(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/problems/nope/like", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
