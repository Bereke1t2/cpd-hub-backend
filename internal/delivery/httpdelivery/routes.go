package httpdelivery

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler defines the handler methods expected by the router.
// Your actual handler implementation (in handler.go or elsewhere)
// should implement these methods: func(c *gin.Context)

// RegisterRoutes registers all API routes under /api.
//   - r is the gin engine
//   - h is your handler implementation
//   - auth is the authorization middleware (will be applied to protected routes).
//     Pass nil if you don't want middleware applied.
//   - loadUser is an optional middleware that loads the authenticated user's profile
//     into the Gin context (stored under key "user"). Pass nil to skip.
func RegisterRoutes(r *gin.Engine, h Handler, auth gin.HandlerFunc, loadUser gin.HandlerFunc) {
	api := r.Group("/api")

	// Strict rate limiter for auth endpoints (5 req/min per IP).
	authLimiter := NewRateLimiter(5, time.Minute)
	// Looser limiter for authenticated writes to keep bursty clients in check.
	writeLimiter := NewRateLimiter(30, time.Minute)

	r.GET("/openapi.yaml", func(c *gin.Context) {
		data, err := os.ReadFile("docs/openapi.yaml")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "openapi spec not available"})
			return
		}
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", data)
	})
	r.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>CPD Hub API Docs</title></head>
<body><main style="font-family:system-ui,sans-serif;max-width:720px;margin:40px auto;line-height:1.5">
<h1>CPD Hub API Docs</h1>
<p>The OpenAPI document is available at <a href="/openapi.yaml">/openapi.yaml</a>.</p>
</main></body>
</html>`))
	})

	// Authentication (no auth middleware)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authLimiter.Middleware(), h.Login)
		authGroup.POST("/signup", authLimiter.Middleware(), h.Signup)
		authGroup.POST("/refresh", authLimiter.Middleware(), h.Refresh) // refresh token in body; no bearer auth
	}

	// Protected: auth required — token validated, user loaded into context.
	protected := api.Group("/")
	if auth != nil {
		protected.Use(auth)
	}
	if loadUser != nil {
		protected.Use(loadUser)
	}

	// Authenticated auth routes (/me)
	protected.GET("/auth/me", h.Me)

	// Problems
	protected.GET("/problems", h.GetProblems)
	protected.GET("/problems/daily", h.GetDailyProblem)
	protected.GET("/problems/:id", h.GetProblem)
	protected.POST("/problems/:id/like", writeLimiter.Middleware(), h.LikeProblem)
	protected.POST("/problems/:id/dislike", writeLimiter.Middleware(), h.DislikeProblem)
	protected.POST("/problems/:id/solve", writeLimiter.Middleware(), h.SolveProblem)
	protected.DELETE("/problems/:id/solve", writeLimiter.Middleware(), h.UnsolveProblem)

	// Contests
	protected.GET("/contests", h.GetContests)
	protected.POST("/contests/:id/participate", writeLimiter.Middleware(), h.ParticipateContest)
	protected.DELETE("/contests/:id/participate", writeLimiter.Middleware(), h.UnparticipateContest)
	protected.GET("/contests/:id/leaderboard", h.GetContestLeaderboard)

	// Users & Profiles
	protected.GET("/users", h.GetUsers)
	protected.GET("/users/profile/:username", h.GetUserProfile)
	protected.GET("/users/profile/:username/analytics/heatmap", h.GetHeatmap)
	protected.GET("/users/profile/:username/analytics/rating-history", h.GetRatingHistory)
	protected.GET("/users/profile/:username/attendance", h.GetAttendance)
	protected.GET("/users/profile/:username/submissions", h.GetSubmissions)

	// Activity & Info
	protected.GET("/activity", h.GetActivity)
	protected.GET("/info", h.GetInfo)

	// Bookmarks
	bm := protected.Group("/bookmarks")
	bm.GET("", h.ListBookmarks)
	bm.POST("/:problemId", writeLimiter.Middleware(), h.AddBookmark)
	bm.DELETE("/:problemId", writeLimiter.Middleware(), h.RemoveBookmark)

	cons := protected.Group("/consistency")
	{
		cons.GET("/streak", h.GetStreak)
		cons.PUT("/streak", writeLimiter.Middleware(), h.PutStreak)
		cons.GET("/goal", h.GetGoal)
		cons.PUT("/goal", writeLimiter.Middleware(), h.PutGoal)
		cons.GET("/ladders", h.GetLadders)
	}

	// Learning
	learn := protected.Group("/learning")
	{
		learn.GET("/topics", h.GetTopics)
		learn.GET("/tracks", h.GetTracks)
		learn.GET("/lessons/:topicId", h.GetLesson)
	}
}
