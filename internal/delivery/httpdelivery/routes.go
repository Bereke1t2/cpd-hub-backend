package httpdelivery

import "github.com/gin-gonic/gin"

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

	// Authentication (no auth middleware)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
		authGroup.POST("/signup", h.Signup)
		authGroup.POST("/refresh", h.Refresh) // refresh token in body; no bearer auth
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
	protected.POST("/problems/:id/like", h.LikeProblem)
	protected.POST("/problems/:id/dislike", h.DislikeProblem)
	protected.POST("/problems/:id/solve", h.SolveProblem)
	protected.DELETE("/problems/:id/solve", h.UnsolveProblem)

	// Contests
	protected.GET("/contests", h.GetContests)
	protected.POST("/contests/:id/participate", h.ParticipateContest)
	protected.DELETE("/contests/:id/participate", h.UnparticipateContest)
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

	cons := protected.Group("/consistency")
	{
		cons.GET("/streak", h.GetStreak)
		cons.PUT("/streak", h.PutStreak)
		cons.GET("/goal", h.GetGoal)
		cons.PUT("/goal", h.PutGoal)
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
