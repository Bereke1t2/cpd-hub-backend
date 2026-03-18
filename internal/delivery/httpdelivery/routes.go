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
	}

	// Protected routes (apply auth middleware if provided)
	protected := api.Group("/")
	if auth != nil {
		protected.Use(auth)
	}
	if loadUser != nil {
		protected.Use(loadUser)
	}

	// Problems
	protected.GET("/problems", h.GetProblems)
	protected.GET("/problems/daily", h.GetDailyProblem)
	protected.POST("/problems/:id/like", h.LikeProblem)
	protected.POST("/problems/:id/dislike", h.DislikeProblem)
	protected.POST("/problems/:id/solve", h.SolveProblem)
	protected.DELETE("/problems/:id/solve", h.UnsolveProblem)

	// Contests
	protected.GET("/contests", h.GetContests)
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
}
