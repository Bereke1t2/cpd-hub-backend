package httpdelivery

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/external"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/security"
	activityuc "github.com/bereket/cpd-hub-backend/internal/usecase/activity"
	authuc "github.com/bereket/cpd-hub-backend/internal/usecase/auth"
	"github.com/gin-gonic/gin"
)

// Repos contains optional repository implementations that handlers may call.
type Repos struct {
	Auth        domain.AuthRepository
	Problem     domain.ProblemRepository
	Contest     domain.ContestRepository
	Profile     domain.ProfileRepository
	Activity    domain.ActivityRepository
	Bookmark    domain.BookmarkRepository
	Info        domain.InfoRepository
	Consistency domain.ConsistencyRepository
	Learning    domain.LearningRepository
}

// Handler is the interface to be implemented by an HTTP handler.
type Handler interface {
	Router() http.Handler
	Healthz(*gin.Context)
	Readyz(*gin.Context)
	Login(*gin.Context)
	Signup(*gin.Context)
	Me(*gin.Context)
	Refresh(*gin.Context)
	GetProblem(*gin.Context)
	GetProblems(*gin.Context)
	GetDailyProblem(*gin.Context)
	LikeProblem(*gin.Context)
	DislikeProblem(*gin.Context)
	SolveProblem(*gin.Context)
	UnsolveProblem(*gin.Context)

	GetContests(*gin.Context)
	ParticipateContest(*gin.Context)
	UnparticipateContest(*gin.Context)
	GetContestLeaderboard(*gin.Context)

	GetUsers(*gin.Context)
	GetUserProfile(*gin.Context)
	GetHeatmap(*gin.Context)
	GetRatingHistory(*gin.Context)
	GetAttendance(*gin.Context)
	GetSubmissions(*gin.Context)

	GetActivity(*gin.Context)
	GetInfo(*gin.Context)

	ListBookmarks(*gin.Context)
	AddBookmark(*gin.Context)
	RemoveBookmark(*gin.Context)

	GetStreak(*gin.Context)
	PutStreak(*gin.Context)
	GetGoal(*gin.Context)
	PutGoal(*gin.Context)
	GetLadders(*gin.Context)

	GetTopics(*gin.Context)
	GetTracks(*gin.Context)
	GetLesson(*gin.Context)
}

// handlerImpl is the concrete implementation used here.
type handlerImpl struct {
	repos    Repos
	authUC   *authuc.UseCase
	recorder *activityuc.Recorder
	db       *postgres.Client
	router   *gin.Engine
	lbCache  *external.TTLCache
}

// NewHandler creates the handler and registers routes.
func NewHandler(repos Repos, db *postgres.Client, corsOrigins []string) Handler {
	g := gin.New()
	g.Use(RecoveryJSON(), RequestID(), SecurityHeaders(), BodySizeLimit(1<<20))

	authUC := authuc.New(repos.Auth)
	var recorder *activityuc.Recorder
	if repos.Activity != nil {
		recorder = activityuc.NewRecorder(repos.Activity)
	}
	h := &handlerImpl{
		repos:    repos,
		authUC:   authUC,
		recorder: recorder,
		db:       db,
		router:   g,
		lbCache:  external.NewTTLCache(60 * time.Second),
	}

	g.GET("/healthz", h.Healthz)
	g.GET("/readyz", h.Readyz)

	var authMiddleware gin.HandlerFunc
	var loadUser gin.HandlerFunc
	if repos.Auth != nil {
		authMiddleware = security.AuthMiddleware()
	}
	if repos.Profile != nil {
		loadUser = security.LoadUser(repos.Profile)
	}

	RegisterRoutes(g, h, authMiddleware, loadUser)

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

// currentUsername returns the authenticated caller's handle from context.
func currentUsername(c *gin.Context) string {
	v, _ := c.Get("username")
	s, _ := v.(string)
	return s
}

// currentUser returns the authenticated caller's profile from context (may be nil).
func currentUser(c *gin.Context) *domain.UserProfile {
	v, _ := c.Get("user")
	u, _ := v.(*domain.UserProfile)
	return u
}

// --- Interface wrapper methods (delegate to existing handlers) ---
func (h *handlerImpl) Login(c *gin.Context)                 { h.authLogin(c) }
func (h *handlerImpl) Signup(c *gin.Context)                { h.authSignup(c) }
func (h *handlerImpl) Me(c *gin.Context)                    { h.authMe(c) }
func (h *handlerImpl) Refresh(c *gin.Context)               { h.authRefresh(c) }
func (h *handlerImpl) GetProblems(c *gin.Context)           { h.problemsList(c) }
func (h *handlerImpl) GetDailyProblem(c *gin.Context)       { h.problemsDaily(c) }
func (h *handlerImpl) LikeProblem(c *gin.Context)           { h.problemsLike(c) }
func (h *handlerImpl) DislikeProblem(c *gin.Context)        { h.problemsDislike(c) }
func (h *handlerImpl) SolveProblem(c *gin.Context)          { h.problemsSolve(c) }
func (h *handlerImpl) UnsolveProblem(c *gin.Context)        { h.problemsUnsolve(c) }
func (h *handlerImpl) GetContests(c *gin.Context)           { h.contestsList(c) }
func (h *handlerImpl) ParticipateContest(c *gin.Context)    { h.contestsParticipate(c) }
func (h *handlerImpl) UnparticipateContest(c *gin.Context)  { h.contestsUnparticipate(c) }
func (h *handlerImpl) GetContestLeaderboard(c *gin.Context) { h.contestLeaderboard(c) }
func (h *handlerImpl) GetUsers(c *gin.Context)              { h.listUsers(c) }
func (h *handlerImpl) GetUserProfile(c *gin.Context)        { h.getProfile(c) }
func (h *handlerImpl) GetHeatmap(c *gin.Context)            { h.profileHeatmap(c) }
func (h *handlerImpl) GetRatingHistory(c *gin.Context)      { h.profileRatingHistory(c) }
func (h *handlerImpl) GetAttendance(c *gin.Context)         { h.profileAttendance(c) }
func (h *handlerImpl) GetSubmissions(c *gin.Context)        { h.profileSubmissions(c) }
func (h *handlerImpl) GetActivity(c *gin.Context)           { h.activityList(c) }
func (h *handlerImpl) GetInfo(c *gin.Context)               { h.infoList(c) }
func (h *handlerImpl) ListBookmarks(c *gin.Context)         { h.bookmarksList(c) }
func (h *handlerImpl) AddBookmark(c *gin.Context)           { h.bookmarksAdd(c) }
func (h *handlerImpl) RemoveBookmark(c *gin.Context)        { h.bookmarksRemove(c) }

func (h *handlerImpl) GetTopics(c *gin.Context) { h.learningTopics(c) }
func (h *handlerImpl) GetTracks(c *gin.Context) { h.learningTracks(c) }
func (h *handlerImpl) GetLesson(c *gin.Context) { h.learningLesson(c) }

// apiProblem shapes a domain.Problem into the JSON the Flutter client expects.
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
		"numberOfSolvedPeople": p.SolverCount, // real count from user_problems
	}
}

// --- Auth handlers ---
func (h *handlerImpl) authLogin(c *gin.Context) {
	var req domain.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	res, err := h.authUC.Login(&req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, gin.H{
		"token":        res.Token,
		"refreshToken": res.RefreshToken,
		"user": gin.H{
			"username": res.User.Username,
			"fullName": res.User.FullName,
		},
	})
}

func (h *handlerImpl) authSignup(c *gin.Context) {
	var req domain.SignupRequest
	if err := bindJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	res, err := h.authUC.Signup(&req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, gin.H{
		"token":        res.Token,
		"refreshToken": res.RefreshToken,
		"user": gin.H{
			"username": res.User.Username,
			"fullName": res.User.FullName,
		},
	})
}

func (h *handlerImpl) authMe(c *gin.Context) {
	// LoadUser has already hydrated the context; fall back to a profile lookup.
	if u := currentUser(c); u != nil {
		respondOK(c, u)
		return
	}
	username := currentUsername(c)
	if username == "" {
		respondError(c, domain.ErrUnauthorized("not authenticated"))
		return
	}
	p, err := h.repos.Profile.GetProfile(username)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, p)
}

func (h *handlerImpl) authRefresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := bindJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	res, err := h.authUC.Refresh(req.RefreshToken)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, gin.H{
		"token":        res.Token,
		"refreshToken": res.RefreshToken,
	})
}

// --- Problems ---
func (h *handlerImpl) problemsList(c *gin.Context) {
	pg := parsePage(c)
	list, err := h.repos.Problem.ListForUser(currentUsername(c), pg.Limit, pg.Offset)
	if err != nil {
		respondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, p := range list {
		out = append(out, apiProblem(p))
	}
	respondOK(c, out)
}

func (h *handlerImpl) problemsDaily(c *gin.Context) {
	p, err := h.repos.Problem.GetDailyForUser(currentUsername(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, apiProblem(p))
}

func (h *handlerImpl) problemsLike(c *gin.Context) {
	username := currentUsername(c)
	id := c.Param("id")
	if err := h.repos.Problem.Like(username, id); err != nil {
		respondError(c, err)
		return
	}
	if h.recorder != nil {
		if p, err := h.repos.Problem.GetByIDForUser(username, id); err == nil {
			h.recorder.RecordLike(username, p.Title, time.Now())
		}
	}
	respondSuccess(c)
}

func (h *handlerImpl) problemsDislike(c *gin.Context) {
	if err := h.repos.Problem.Dislike(currentUsername(c), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c)
}

func (h *handlerImpl) problemsSolve(c *gin.Context) {
	username := currentUsername(c)
	id := c.Param("id")
	if err := h.repos.Problem.MarkSolved(username, id); err != nil {
		respondError(c, err)
		return
	}
	if h.recorder != nil {
		if p, err := h.repos.Problem.GetByIDForUser(username, id); err == nil {
			h.recorder.RecordSolve(username, p.Title, time.Now())
		}
	}
	respondSuccess(c)
}

func (h *handlerImpl) problemsUnsolve(c *gin.Context) {
	if err := h.repos.Problem.UnmarkSolved(currentUsername(c), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c)
}

// --- Contests ---
func (h *handlerImpl) contestsList(c *gin.Context) {
	list, err := h.repos.Contest.ListForUser(currentUsername(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, list)
}

func (h *handlerImpl) contestsParticipate(c *gin.Context) {
	if err := h.repos.Contest.Participate(currentUsername(c), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c)
}

func (h *handlerImpl) contestsUnparticipate(c *gin.Context) {
	if err := h.repos.Contest.Unparticipate(currentUsername(c), c.Param("id")); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c)
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

		// leaderboard caching
		cacheKey := fmt.Sprintf("lb:%d:%d:%d:%v", contestID, fromI, countI, showUnofficial)
		if v, fresh, ok := h.lbCache.Get(cacheKey); ok && fresh {
			respondOK(c, v)
			return
		}

		client := external.NewKontestsClient()
		rows, _, err := client.FetchContestStandings(contestID, fromI, countI, showUnofficial)
		if err != nil {
			// fallback to stale cache if available
			if v, _, ok := h.lbCache.Get(cacheKey); ok {
				respondOK(c, v)
				return
			}
			respondError(c, domain.ErrInternal("could not fetch standings").Wrap(err))
			return
		}

		// map to domain.LeaderboardEntry
		out := make([]*domain.LeaderboardEntry, 0, len(rows))
		for _, r := range rows {
			out = append(out, &domain.LeaderboardEntry{
				Rank:        r.Rank,
				Username:    r.Handle,
				Rating:      r.Rating,
				Score:       int(r.Points),
				Penalty:     r.Penalty,
				SolvedCount: r.Solved,
			})
		}
		h.lbCache.Set(cacheKey, out)
		respondOK(c, out)
		return
	}

	// Fallback to repo-provided leaderboard
	lb, err := h.repos.Contest.Leaderboard(id)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, lb)
}

// --- Profiles / Users ---
func (h *handlerImpl) listUsers(c *gin.Context) {
	pg := parsePage(c)
	list, err := h.repos.Profile.ListUsers(pg.Limit, pg.Offset)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, list)
}

func (h *handlerImpl) getProfile(c *gin.Context) {
	username := c.Param("username")
	p, err := h.repos.Profile.GetProfile(username)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, p)
}

func (h *handlerImpl) profileHeatmap(c *gin.Context) {
	username := c.Param("username")
	hm, err := h.repos.Profile.GetProfileHeatmap(username)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, hm)
}

func (h *handlerImpl) profileRatingHistory(c *gin.Context) {
	username := c.Param("username")
	rh, err := h.repos.Profile.GetProfileRatingHistory(username)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, rh)
}

func (h *handlerImpl) profileAttendance(c *gin.Context) {
	username := c.Param("username")
	att, err := h.repos.Profile.GetProfileAttendance(username)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, att)
}

func (h *handlerImpl) profileSubmissions(c *gin.Context) {
	username := c.Param("username")
	subs, err := h.repos.Profile.GetProfileSubmissions(username)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, subs)
}

// --- Activity & Info ---
func (h *handlerImpl) activityList(c *gin.Context) {
	pg := parsePage(c)
	list, err := h.repos.Activity.List(pg.Limit, pg.Offset)
	if err != nil {
		respondError(c, err)
		return
	}
	now := time.Now()
	for _, a := range list {
		a.Timestamp = activityuc.HumanizeSince(a.Timestamp, now)
	}
	respondOK(c, list)
}

func (h *handlerImpl) infoList(c *gin.Context) {
	list, err := h.repos.Info.List()
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, list)
}

// --- Bookmarks ---
func (h *handlerImpl) bookmarksList(c *gin.Context) {
	username := currentUsername(c)
	ids, err := h.repos.Bookmark.ListProblemIDs(username)
	if err != nil {
		respondError(c, err)
		return
	}
	out := make([]gin.H, 0, len(ids))
	for _, id := range ids {
		if p, err := h.repos.Problem.GetByIDForUser(username, id); err == nil {
			out = append(out, apiProblem(p))
		}
	}
	respondOK(c, out)
}

func (h *handlerImpl) bookmarksAdd(c *gin.Context) {
	if err := h.repos.Bookmark.Add(currentUsername(c), c.Param("problemId")); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c)
}

func (h *handlerImpl) bookmarksRemove(c *gin.Context) {
	if err := h.repos.Bookmark.Remove(currentUsername(c), c.Param("problemId")); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c)
}

func (h *handlerImpl) GetProblem(c *gin.Context) {
	p, err := h.repos.Problem.GetByIDForUser(currentUsername(c), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, apiProblem(p))
}
