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
