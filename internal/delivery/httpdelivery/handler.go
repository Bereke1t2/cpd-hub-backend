package httpdelivery

import (
	"net/http"
	"strconv"
	"strings"

	// "time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/external"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/security"
	contestsuc "github.com/bereket/cpd-hub-backend/internal/usecase/contests"
	"github.com/gin-gonic/gin"
)

// Repos contains optional repository implementations that handlers may call.
type Repos struct {
	User     domain.UserRepository
	Auth     domain.AuthRepository
	Problem  domain.ProblemRepository
	Contest  domain.ContestRepository
	Profile  domain.ProfileRepository
	Activity domain.ActivityRepository
	Info     domain.InfoRepository
}

// Handler is the interface to be implemented by an HTTP handler.
type Handler interface {
	Router() http.Handler
	Login(*gin.Context)
	Signup(*gin.Context)
	GetProblem(*gin.Context)
	GetProblems(*gin.Context)
	GetDailyProblem(*gin.Context)
	LikeProblem(*gin.Context)
	DislikeProblem(*gin.Context)
	SolveProblem(*gin.Context)
	UnsolveProblem(*gin.Context)

	GetContests(*gin.Context)
	GetContestLeaderboard(*gin.Context)

	GetUsers(*gin.Context)
	GetUserProfile(*gin.Context)
	GetHeatmap(*gin.Context)
	GetRatingHistory(*gin.Context)
	GetAttendance(*gin.Context)
	GetSubmissions(*gin.Context)

	GetActivity(*gin.Context)
	GetInfo(*gin.Context)
}

// handlerImpl is the concrete implementation used here.
type handlerImpl struct {
	repos  Repos
	router *gin.Engine
}

// NewHandler creates the handler and registers routes via RegisterRoutes (centralized in routes.go).
func NewHandler(repos Repos) Handler {
	g := gin.Default()
	h := &handlerImpl{repos: repos, router: g}

	// If an Auth repository is provided, enable JWT auth middleware for protected routes.
	var authMiddleware gin.HandlerFunc
	if repos.Auth != nil {
		authMiddleware = security.AuthMiddleware()
	}

	// Register all routes in a single place (routes.go).
	RegisterRoutes(g, h, authMiddleware, nil)

	return h
}

// Router returns the underlying Gin engine (implements http.Handler).
func (h *handlerImpl) Router() http.Handler {
	return h.router
}

// Engine returns the underlying Gin engine for advanced use.
func (h *handlerImpl) Engine() *gin.Engine {
	return h.router
}

// --- Interface wrapper methods (delegate to existing handlers) ---
func (h *handlerImpl) Login(c *gin.Context)                 { h.authLogin(c) }
func (h *handlerImpl) Signup(c *gin.Context)                { h.authSignup(c) }
func (h *handlerImpl) GetProblems(c *gin.Context)           { h.problemsList(c) }
func (h *handlerImpl) GetDailyProblem(c *gin.Context)       { h.problemsDaily(c) }
func (h *handlerImpl) LikeProblem(c *gin.Context)           { h.problemsLike(c) }
func (h *handlerImpl) DislikeProblem(c *gin.Context)        { h.problemsDislike(c) }
func (h *handlerImpl) SolveProblem(c *gin.Context)          { h.problemsSolve(c) }
func (h *handlerImpl) UnsolveProblem(c *gin.Context)        { h.problemsUnsolve(c) }
func (h *handlerImpl) GetContests(c *gin.Context)           { h.contestsList(c) }
func (h *handlerImpl) GetContestLeaderboard(c *gin.Context) { h.contestLeaderboard(c) }
func (h *handlerImpl) GetUsers(c *gin.Context)              { h.listUsers(c) }
func (h *handlerImpl) GetUserProfile(c *gin.Context)        { h.getProfile(c) }
func (h *handlerImpl) GetHeatmap(c *gin.Context)            { h.profileHeatmap(c) }
func (h *handlerImpl) GetRatingHistory(c *gin.Context)      { h.profileRatingHistory(c) }
func (h *handlerImpl) GetAttendance(c *gin.Context)         { h.profileAttendance(c) }
func (h *handlerImpl) GetSubmissions(c *gin.Context)        { h.profileSubmissions(c) }
func (h *handlerImpl) GetActivity(c *gin.Context)           { h.activityList(c) }
func (h *handlerImpl) GetInfo(c *gin.Context)               { h.infoList(c) }

// helper: shape domain.Problem into API-friendly map with aliases expected by client
func apiProblem(p *domain.Problem) gin.H {
	return gin.H{
		"id":                   p.ID,
		"problemId":            p.ID,
		"title":                p.Title,
		"difficulty":           p.Difficulty,
		"topicTags":            p.TopicTags,
		"numberOfLikes":        p.Likes,
		"numberOfDislikes":     p.Dislikes,
		"problemUrl":           p.DeepLink,
		"isLiked":              p.IsLiked,
		"isDisliked":           p.IsDisliked,
		"solved":               p.Solved,
		"numberOfSolvedPeople": 0, // not tracked yet in domain; placeholder
	}
}

// --- Auth handlers ---
func (h *handlerImpl) authLogin(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": "bad json"})
		return
	}
	if h.repos.Auth != nil {
		res, err := h.repos.Auth.Login(&req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": err.Error()})
			return
		}
		// shape response to include email when available
		userMap := gin.H{"username": res.User.Username, "fullName": res.User.FullName}
		if req.Email != "" {
			userMap["email"] = req.Email
		}
		c.JSON(http.StatusOK, gin.H{"token": res.Token, "user": userMap})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": "sample-token", "user": gin.H{"username": "bereket", "fullName": "Bereket Lemma", "email": "test@example.com"}})
}

func (h *handlerImpl) authSignup(c *gin.Context) {
	var req domain.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": "bad json"})
		return
	}
	if h.repos.Auth != nil {
		res, err := h.repos.Auth.Signup(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "could not signup", "message": err.Error()})
			return
		}
		userMap := gin.H{"username": res.User.Username, "fullName": res.User.FullName}
		if req.Email != "" {
			userMap["email"] = req.Email
		}
		c.JSON(http.StatusCreated, gin.H{"token": res.Token, "user": userMap})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": "sample-token", "user": gin.H{"username": "newuser", "fullName": req.FullName, "email": req.Email}})
}

// --- Problems ---
func (h *handlerImpl) problemsList(c *gin.Context) {
	if h.repos.Problem != nil {
		list, err := h.repos.Problem.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list problems", "message": err.Error()})
			return
		}
		out := make([]gin.H, 0, len(list))
		for _, p := range list {
			out = append(out, apiProblem(p))
		}
		c.JSON(http.StatusOK, out)
		return
	}
	c.JSON(http.StatusOK, []gin.H{apiProblem(&domain.Problem{ID: "p1", Title: "Two Sum", Difficulty: "Easy", TopicTags: []string{"Array", "Hash Table"}, Likes: 245, Dislikes: 12, DeepLink: "https://...", IsLiked: false, IsDisliked: false, Solved: true})})
}

func (h *handlerImpl) problemsDaily(c *gin.Context) {
	if h.repos.Problem != nil {
		p, err := h.repos.Problem.GetDaily()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get daily problem", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, apiProblem(p))
		return
	}
	c.JSON(http.StatusOK, apiProblem(&domain.Problem{ID: "dp1", Title: "Longest Common Subsequence", Difficulty: "Medium", TopicTags: []string{"Dynamic Programming", "String"}, Likes: 342, Dislikes: 18, DeepLink: "https://...", IsLiked: false, IsDisliked: false, Solved: false}))
}

func (h *handlerImpl) problemsLike(c *gin.Context) {
	id := c.Param("id")
	if h.repos.Auth != nil {
		if _, ok := security.GetClaims(c); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing or invalid token"})
			return
		}
	}
	if h.repos.Problem != nil {
		// verify existence
		list, err := h.repos.Problem.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list problems", "message": err.Error()})
			return
		}
		found := false
		for _, p := range list {
			if p.ID == id {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
			return
		}
		if err := h.repos.Problem.Like(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not like", "message": err.Error(), "problemId": id})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *handlerImpl) problemsDislike(c *gin.Context) {
	id := c.Param("id")
	if h.repos.Auth != nil {
		if _, ok := security.GetClaims(c); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing or invalid token"})
			return
		}
	}
	if h.repos.Problem != nil {
		list, err := h.repos.Problem.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list problems", "message": err.Error()})
			return
		}
		found := false
		for _, p := range list {
			if p.ID == id {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
			return
		}
		if err := h.repos.Problem.Dislike(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not dislike", "message": err.Error(), "problemId": id})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *handlerImpl) problemsSolve(c *gin.Context) {
	id := c.Param("id")
	if h.repos.Auth != nil {
		if _, ok := security.GetClaims(c); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing or invalid token"})
			return
		}
	}
	if h.repos.Problem != nil {
		list, err := h.repos.Problem.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list problems", "message": err.Error()})
			return
		}
		found := false
		for _, p := range list {
			if p.ID == id {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
			return
		}
		if err := h.repos.Problem.MarkSolved(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not mark solved", "message": err.Error(), "problemId": id})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *handlerImpl) problemsUnsolve(c *gin.Context) {
	id := c.Param("id")
	if h.repos.Auth != nil {
		if _, ok := security.GetClaims(c); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing or invalid token"})
			return
		}
	}
	if h.repos.Problem != nil {
		list, err := h.repos.Problem.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list problems", "message": err.Error()})
			return
		}
		found := false
		for _, p := range list {
			if p.ID == id {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
			return
		}
		if err := h.repos.Problem.UnmarkSolved(id); err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found", "message": "problem not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not unmark solved", "message": err.Error(), "problemId": id})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Contests ---
func (h *handlerImpl) contestsList(c *gin.Context) {
	// create kontests client and usecase to fetch platform contests and merge with repo
	client := external.NewKontestsClient()
	uc := contestsuc.NewWithClient(h.repos.Contest, client)
	list, err := uc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list contests", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *handlerImpl) contestLeaderboard(c *gin.Context) {
	id := c.Param("id")
	// If this looks like a Codeforces contest id produced by our Fetch (e.g. "codeforces-1932"), fetch standings from Codeforces API.
	if strings.HasPrefix(strings.ToLower(id), "codeforces-") {
		parts := strings.SplitN(id, "-", 2)
		if len(parts) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest id"})
			return
		}
		contestIDStr := parts[1]
		contestID, err := strconv.Atoi(contestIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid codeforces contest id", "message": err.Error()})
			return
		}

		// optional query params: from, count, showUnofficial
		fromStr := c.DefaultQuery("from", "1")
		countStr := c.DefaultQuery("count", "50")
		showUnofficialStr := c.DefaultQuery("showUnofficial", "false")
		fromI, _ := strconv.Atoi(fromStr)
		countI, _ := strconv.Atoi(countStr)
		showUnofficial := strings.EqualFold(showUnofficialStr, "true")

		client := external.NewKontestsClient()
		rows, _, err := client.FetchContestStandings(contestID, fromI, countI, showUnofficial)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch standings from codeforces", "message": err.Error()})
			return
		}

		// map to domain.LeaderboardEntry
		out := make([]*domain.LeaderboardEntry, 0, len(rows))
		for _, r := range rows {
			out = append(out, &domain.LeaderboardEntry{
				Rank:           r.Rank,
				Username:       r.Handle,
				Rating:         0,
				Score:          int(r.Points),
				Penalty:        r.Penalty,
				ProblemsSolved: []string{},
			})
		}
		c.JSON(http.StatusOK, out)
		return
	}

	// Fallback to repo-provided leaderboard
	if h.repos.Contest != nil {
		lb, err := h.repos.Contest.Leaderboard(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get leaderboard"})
			return
		}
		c.JSON(http.StatusOK, lb)
		return
	}
	c.JSON(http.StatusOK, []domain.LeaderboardEntry{{Rank: 1, Username: "tourist", Rating: 3800, Score: 600, Penalty: 45, ProblemsSolved: []string{"A", "B", "C", "D", "E", "F"}}})
}

// --- Profiles / Users ---
func (h *handlerImpl) listUsers(c *gin.Context) {
	if h.repos.Profile != nil {
		list, err := h.repos.Profile.ListUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list users"})
			return
		}
		c.JSON(http.StatusOK, list)
		return
	}
	c.JSON(http.StatusOK, []domain.UserProfile{{Username: "bereket", FullName: "Bereket Lemma", Bio: "Competitive programmer", AvatarURL: "https://...", Rating: 1750}})
}

func (h *handlerImpl) getProfile(c *gin.Context) {
	username := c.Param("username")
	if h.repos.Profile != nil {
		p, err := h.repos.Profile.GetProfile(username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, p)
		return
	}
	c.JSON(http.StatusOK, domain.UserProfile{Username: username, FullName: "Bereket Lemma", Bio: "Competitive programmer | CPD Hub enthusiast"})
}

func (h *handlerImpl) profileHeatmap(c *gin.Context) {
	c.JSON(http.StatusOK, []map[string]interface{}{{"date": "2026-02-01", "solveCount": 0}, {"date": "2026-02-02", "solveCount": 3}})
}

func (h *handlerImpl) profileRatingHistory(c *gin.Context) {
	c.JSON(http.StatusOK, []map[string]interface{}{{"date": "2025-08-01", "rating": 1000}, {"date": "2026-01-01", "rating": 1750}})
}

func (h *handlerImpl) profileAttendance(c *gin.Context) {
	c.JSON(http.StatusOK, []map[string]string{{"date": "2026-02-01", "status": "Present"}, {"date": "2026-02-02", "status": "Absent"}})
}

func (h *handlerImpl) profileSubmissions(c *gin.Context) {
	c.JSON(http.StatusOK, []map[string]interface{}{{"id": "s1", "problemId": "p1", "problemTitle": "Two Sum", "status": "Accepted", "language": "Python", "executionTime": "45ms", "memoryUsed": "14.2MB", "timestamp": "2 hours ago"}})
}

// --- Activity & Info ---
func (h *handlerImpl) activityList(c *gin.Context) {
	if h.repos.Activity != nil {
		list, err := h.repos.Activity.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list activity"})
			return
		}
		c.JSON(http.StatusOK, list)
		return
	}
	c.JSON(http.StatusOK, []domain.Activity{{ID: "a1", Username: "abel", Action: "solved 'Two Sum' in 3 min", Type: "Solve", Timestamp: "2 min ago"}})
}

func (h *handlerImpl) infoList(c *gin.Context) {
	if h.repos.Info != nil {
		list, err := h.repos.Info.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list info"})
			return
		}
		c.JSON(http.StatusOK, list)
		return
	}
	c.JSON(http.StatusOK, []domain.Info{{Title: "System Maintenance", Description: "Scheduled maintenance on Feb 20th from 2-4 AM"}})
}

func (h *handlerImpl) GetProblem(c *gin.Context) {
	id := c.Param("id")
	if h.repos.Problem != nil {
		p, err := h.repos.Problem.GetById(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, apiProblem(p))
		return
	}
	c.JSON(http.StatusOK, apiProblem(&domain.Problem{ID: id, Title: "Sample Problem", Difficulty: "Medium", TopicTags: []string{"Example"}, Likes: 100, Dislikes: 5, DeepLink: "https://...", IsLiked: false, IsDisliked: false, Solved: false}))
}
