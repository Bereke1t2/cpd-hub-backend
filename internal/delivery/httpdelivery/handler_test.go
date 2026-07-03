package httpdelivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func newTestRouter(problemRepo *mockProblemRepo, opts ...func(*Repos)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	repos := Repos{Auth: &mockAuthRepo{}, Problem: problemRepo, Profile: &mockProfileRepo{}}
	for _, opt := range opts {
		opt(&repos)
	}
	h := NewHandler(repos, nil, nil)
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

func TestMetricsEndpointExposesPrometheusMetrics(t *testing.T) {
	router := newTestRouter(&mockProblemRepo{})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "http_requests_total") {
		t.Fatalf("metrics body does not include http_requests_total:\n%s", body)
	}
}

type mockCourseRepo struct {
	completedCourse string
	completedLesson string
}

func (m *mockCourseRepo) List(string) ([]*domain.Course, error) {
	return []*domain.Course{{
		ID:      "cp-foundations",
		Title:   "CP Foundations",
		Summary: "Basics",
		Level:   "Beginner",
		Modules: []domain.CourseModule{{
			ID:    "m1",
			Title: "Start",
			Lessons: []domain.CourseLesson{{
				ID:         "l1",
				Title:      "Fast I/O",
				Kind:       "video",
				ContentURL: "https://example.com",
				Completed:  false,
			}},
		}},
	}}, nil
}

func (m *mockCourseRepo) Get(string, string) (*domain.Course, error) {
	list, _ := m.List("")
	return list[0], nil
}

func (m *mockCourseRepo) CompleteLesson(_ string, courseID, lessonID string) error {
	m.completedCourse = courseID
	m.completedLesson = lessonID
	return nil
}

func TestCoursesListReturnsNestedCourses(t *testing.T) {
	router := newTestRouter(&mockProblemRepo{}, func(r *Repos) {
		r.Course = &mockCourseRepo{}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"contentUrl"`) {
		t.Fatalf("body missing course lesson contentUrl: %s", w.Body.String())
	}
}

func TestCompleteLessonReturnsCompletionPayload(t *testing.T) {
	repo := &mockCourseRepo{}
	router := newTestRouter(&mockProblemRepo{}, func(r *Repos) {
		r.Course = repo
	})

	req := httptest.NewRequest(http.MethodPost, "/api/courses/cp-foundations/lessons/l1/complete", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.completedCourse != "cp-foundations" || repo.completedLesson != "l1" {
		t.Fatalf("completed = %q/%q, want cp-foundations/l1", repo.completedCourse, repo.completedLesson)
	}
	if !strings.Contains(w.Body.String(), `"completed":true`) {
		t.Fatalf("body missing completed=true: %s", w.Body.String())
	}
}

type mockPracticeRepo struct {
	current       *domain.ReviewItem
	updatedReview *domain.ReviewItem
	deleted       string
	upsolve       *domain.UpsolveItem
}

func (m *mockPracticeRepo) ListReviewQueue(string) ([]*domain.ReviewItem, error) {
	return []*domain.ReviewItem{m.current}, nil
}

func (m *mockPracticeRepo) GetReview(string, string) (*domain.ReviewItem, error) {
	if m.current == nil {
		return nil, domain.ErrNotFound("review item not found")
	}
	return m.current, nil
}

func (m *mockPracticeRepo) AddReview(_ string, item *domain.ReviewItem) (*domain.ReviewItem, error) {
	m.updatedReview = item
	return item, nil
}

func (m *mockPracticeRepo) UpdateReview(_ string, item *domain.ReviewItem) (*domain.ReviewItem, error) {
	m.updatedReview = item
	return item, nil
}

func (m *mockPracticeRepo) DeleteReview(_ string, problemID string) error {
	m.deleted = problemID
	return nil
}

func (m *mockPracticeRepo) ListUpsolves(string) ([]*domain.UpsolveItem, error) {
	if m.upsolve == nil {
		return []*domain.UpsolveItem{}, nil
	}
	return []*domain.UpsolveItem{m.upsolve}, nil
}

func (m *mockPracticeRepo) AddUpsolve(_ string, item *domain.UpsolveItem) (*domain.UpsolveItem, error) {
	m.upsolve = item
	return item, nil
}

func (m *mockPracticeRepo) UpdateUpsolve(_ string, problemID string, resolved bool) (*domain.UpsolveItem, error) {
	m.upsolve = &domain.UpsolveItem{ProblemID: problemID, Resolved: resolved}
	return m.upsolve, nil
}

func TestReviewUpdateWithQualityUsesStoredCard(t *testing.T) {
	repo := &mockPracticeRepo{current: &domain.ReviewItem{
		ProblemID:   "p1",
		DueDate:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Interval:    6,
		Ease:        2.5,
		Repetitions: 2,
	}}
	router := newTestRouter(&mockProblemRepo{}, func(r *Repos) {
		r.Practice = repo
	})

	req := httptest.NewRequest(http.MethodPut, "/api/practice/review-queue/p1", strings.NewReader(`{"quality":5}`))
	req.Header.Set("Authorization", authHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	if repo.updatedReview == nil || repo.updatedReview.Interval <= 6 || repo.updatedReview.Repetitions != 3 {
		t.Fatalf("updated review = %+v, want interval grown from stored card", repo.updatedReview)
	}
}

func TestDeleteReviewReturns204(t *testing.T) {
	repo := &mockPracticeRepo{}
	router := newTestRouter(&mockProblemRepo{}, func(r *Repos) {
		r.Practice = repo
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/practice/review-queue/p1", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if repo.deleted != "p1" {
		t.Fatalf("deleted = %q, want p1", repo.deleted)
	}
}

type mockArticleRepo struct {
	filter domain.ArticleFilter
}

func (m *mockArticleRepo) List(filter domain.ArticleFilter) ([]*domain.Article, error) {
	m.filter = filter
	return []*domain.Article{{
		ID:          "art-dp",
		Title:       "DP States",
		Author:      "CPD Hub",
		Source:      "cpdhub",
		SourceURL:   "https://example.com",
		Excerpt:     "DP",
		FullContent: "Full",
		PublishedAt: time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Tags:        []string{"dp"},
		Rating:      1500,
	}}, nil
}

func TestArticlesPassesFiltersAndReturnsCamelCase(t *testing.T) {
	repo := &mockArticleRepo{}
	router := newTestRouter(&mockProblemRepo{}, func(r *Repos) {
		r.Article = repo
	})

	req := httptest.NewRequest(http.MethodGet, "/api/articles?limit=250&offset=2&source=cpdhub&tag=dp", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.filter.Limit != 100 || repo.filter.Offset != 2 || repo.filter.Source != "cpdhub" || repo.filter.Tag != "dp" {
		t.Fatalf("filter = %+v, want clamped limit/source/tag", repo.filter)
	}
	var articles []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &articles); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if len(articles) != 1 || articles[0]["sourceUrl"] == nil || articles[0]["fullContent"] == nil {
		t.Fatalf("article shape = %+v, want camelCase sourceUrl/fullContent", articles)
	}
}

func TestArticlesDefaultLimitIsTen(t *testing.T) {
	repo := &mockArticleRepo{}
	router := newTestRouter(&mockProblemRepo{}, func(r *Repos) {
		r.Article = repo
	})

	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.filter.Limit != 10 {
		t.Fatalf("limit = %d, want 10", repo.filter.Limit)
	}
}

func TestArticlesInvalidLimitFallsBackToTen(t *testing.T) {
	repo := &mockArticleRepo{}
	router := newTestRouter(&mockProblemRepo{}, func(r *Repos) {
		r.Article = repo
	})

	req := httptest.NewRequest(http.MethodGet, "/api/articles?limit=0", nil)
	req.Header.Set("Authorization", authHeader(t))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.filter.Limit != 10 {
		t.Fatalf("limit = %d, want 10", repo.filter.Limit)
	}
}
