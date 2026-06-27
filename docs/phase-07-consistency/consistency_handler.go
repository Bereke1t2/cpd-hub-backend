//go:build ignore

// Template for Phase 7 — copy to: internal/delivery/httpdelivery/consistency_handler.go
//
// Thin handlers delegating to the consistency usecase. currentUsername(c) comes
// from the loadUser middleware (Phase 3).
//
package httpdelivery

import (
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/usecase/consistency"
	"github.com/gin-gonic/gin"
)

func (h *handlerImpl) consistencyUC() *consistency.UseCase {
	return consistency.New(h.repos.Consistency)
}

func (h *handlerImpl) GetStreak(c *gin.Context) {
	s, err := h.consistencyUC().GetStreak(currentUsername(c), time.Now().UTC())
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, s)
}

func (h *handlerImpl) PutStreak(c *gin.Context) {
	var s domain.Streak
	if err := bindJSON(c, &s); err != nil {
		respondError(c, err)
		return
	}
	if err := h.repos.Consistency.SaveStreak(currentUsername(c), &s); err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, s)
}

func (h *handlerImpl) GetGoal(c *gin.Context) {
	g, err := h.consistencyUC().GetGoal(currentUsername(c), time.Now().UTC())
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, g)
}

func (h *handlerImpl) PutGoal(c *gin.Context) {
	var g domain.Goal
	if err := bindJSON(c, &g); err != nil {
		respondError(c, err)
		return
	}
	saved, err := h.consistencyUC().SaveGoal(currentUsername(c), &g)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, saved)
}

func (h *handlerImpl) GetLadders(c *gin.Context) {
	ls, err := h.consistencyUC().GetLadders(currentUsername(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, ls)
}

// --- routes (add to RegisterRoutes, inside the protected group) ---
//
//	cons := protected.Group("/consistency")
//	cons.GET("/streak", h.GetStreak)
//	cons.PUT("/streak", h.PutStreak)
//	cons.GET("/goal", h.GetGoal)
//	cons.PUT("/goal", h.PutGoal)
//	cons.GET("/ladders", h.GetLadders)
//
// --- also: add `Consistency domain.ConsistencyRepository` to Repos and the
//     GetStreak/PutStreak/... methods to the Handler interface, and wire
//     `Consistency: databases.NewConsistencyRepositoryDB(client)` in main.go ---
