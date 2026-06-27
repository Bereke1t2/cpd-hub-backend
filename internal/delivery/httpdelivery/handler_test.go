package httpdelivery

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/security"
	"github.com/gin-gonic/gin"
)

type mockProblemRepo struct {
	likeErr  error
	solveErr error
	getErr   error
	liked    string
	solved   string
}

func (m *mockProblemRepo) ListForUser(string, int, int) ([]*domain.Problem, error) { return nil, nil }
func (m *mockProblemRepo) GetByIDForUser(_ string, id string) (*domain.Problem, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return &domain.Problem{ID: id, Title: "Two Sum"}, nil
}
func (m *mockProblemRepo) GetDailyForUser(string) (*domain.Problem, error) {
	return &domain.Problem{}, nil
}
func (m *mockProblemRepo) Like(_ string, id string) error       { m.liked = id; return m.likeErr }
func (m *mockProblemRepo) Dislike(string, string) error         { return nil }
func (m *mockProblemRepo) MarkSolved(_ string, id string) error { m.solved = id; return m.solveErr }
func (m *mockProblemRepo) UnmarkSolved(string, string) error    { return nil }
func (m *mockProblemRepo) CountSolvers(string) (int, error)     { return 0, nil }

type mockAuthRepo struct{}

func (m *mockAuthRepo) FindByEmailOrUsername(string) (*domain.UserRecord, error) {
	return nil, domain.ErrNotFound("not used")
}
func (m *mockAuthRepo) ExistsEmail(string) (bool, error)   { return false, nil }
func (m *mockAuthRepo) UsernameTaken(string) (bool, error) { return false, nil }
func (m *mockAuthRepo) Insert(*domain.UserRecord) error    { return nil }

type mockProfileRepo struct{}

func (m *mockProfileRepo) ListUsers(int, int) ([]*domain.UserProfile, error) { return nil, nil }
func (m *mockProfileRepo) GetProfile(username string) (*domain.UserProfile, error) {
	return &domain.UserProfile{Username: username, FullName: "Alice Example"}, nil
}
func (m *mockProfileRepo) CreateUser(*domain.UserProfile) error                    { return nil }
func (m *mockProfileRepo) UpdateUser(*domain.UserProfile) error                    { return nil }
func (m *mockProfileRepo) DeleteUser(string) error                                 { return nil }
func (m *mockProfileRepo) GetProfileHeatmap(string) ([]domain.HeatmapEntry, error) { return nil, nil }
func (m *mockProfileRepo) GetProfileRatingHistory(string) ([]domain.RatingEntry, error) {
	return nil, nil
}
func (m *mockProfileRepo) GetProfileAttendance(string) ([]domain.AttendanceEntry, error) {
	return nil, nil
}
func (m *mockProfileRepo) GetProfileSubmissions(string) ([]domain.Submission, error) { return nil, nil }

func newTestRouter(problemRepo *mockProblemRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Repos{Auth: &mockAuthRepo{}, Problem: problemRepo, Profile: &mockProfileRepo{}}, nil, nil)
	_ = h
	return h.(*handlerImpl).router
}

func authHeader(t *testing.T) string {
	t.Helper()
	token, err := security.GenerateToken(&domain.UserProfile{Username: "alice", FullName: "Alice Example"}, "alice@example.com", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	return "Bearer " + token
}

func TestLikeProblem_Returns200(t *testing.T) {
	repo := &mockProblemRepo{}
	router := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/problems/p1/like", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.liked != "p1" {
		t.Fatalf("liked = %q, want p1", repo.liked)
	}
}

func TestSolveProblem_Returns200(t *testing.T) {
	repo := &mockProblemRepo{}
	router := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/problems/p1/solve", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.solved != "p1" {
		t.Fatalf("solved = %q, want p1", repo.solved)
	}
}

func TestLikeProblem_NotFound(t *testing.T) {
	repo := &mockProblemRepo{likeErr: domain.ErrNotFound("problem not found")}
	router := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/problems/nope/like", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestLikeProblem_Unauthorized(t *testing.T) {
	repo := &mockProblemRepo{}
	router := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/problems/p1/like", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
