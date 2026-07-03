package httpdelivery

import (
	"net/http"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	practiceuc "github.com/bereket/cpd-hub-backend/internal/usecase/practice"
	"github.com/gin-gonic/gin"
)

type reviewUpdateRequest struct {
	domain.ReviewItem
	Quality *int `json:"quality"`
}

type upsolveUpdateRequest struct {
	Resolved bool `json:"resolved"`
}

func (h *handlerImpl) practiceListReviewQueue(c *gin.Context) {
	if h.repos.Practice == nil {
		respondError(c, domain.ErrInternal("practice repository unavailable"))
		return
	}
	items, err := h.repos.Practice.ListReviewQueue(currentUsername(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, items)
}

func (h *handlerImpl) practiceAddReviewItem(c *gin.Context) {
	if h.repos.Practice == nil {
		respondError(c, domain.ErrInternal("practice repository unavailable"))
		return
	}
	var item domain.ReviewItem
	if err := bindJSON(c, &item); err != nil {
		respondError(c, err)
		return
	}
	if item.ProblemID == "" {
		respondError(c, domain.ErrValidation("problem_id is required"))
		return
	}
	applyNewReviewDefaults(&item)
	saved, err := h.repos.Practice.AddReview(currentUsername(c), &item)
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, saved)
}

func (h *handlerImpl) practiceUpdateReviewItem(c *gin.Context) {
	if h.repos.Practice == nil {
		respondError(c, domain.ErrInternal("practice repository unavailable"))
		return
	}
	var req reviewUpdateRequest
	if err := bindJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}

	username := currentUsername(c)
	problemID := c.Param("problemId")
	item := req.ReviewItem
	item.ProblemID = problemID
	if req.Quality != nil {
		current, err := h.repos.Practice.GetReview(username, problemID)
		if err != nil {
			respondError(c, err)
			return
		}
		item = *practiceuc.Schedule(current, *req.Quality, time.Now())
		item.ProblemID = problemID
	} else {
		applyNewReviewDefaults(&item)
	}

	saved, err := h.repos.Practice.UpdateReview(username, &item)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, saved)
}

func (h *handlerImpl) practiceDeleteReviewItem(c *gin.Context) {
	if h.repos.Practice == nil {
		respondError(c, domain.ErrInternal("practice repository unavailable"))
		return
	}
	if err := h.repos.Practice.DeleteReview(currentUsername(c), c.Param("problemId")); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *handlerImpl) practiceListUpsolves(c *gin.Context) {
	if h.repos.Practice == nil {
		respondError(c, domain.ErrInternal("practice repository unavailable"))
		return
	}
	items, err := h.repos.Practice.ListUpsolves(currentUsername(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, items)
}

func (h *handlerImpl) practiceAddUpsolve(c *gin.Context) {
	if h.repos.Practice == nil {
		respondError(c, domain.ErrInternal("practice repository unavailable"))
		return
	}
	var item domain.UpsolveItem
	if err := bindJSON(c, &item); err != nil {
		respondError(c, err)
		return
	}
	if item.ProblemID == "" {
		respondError(c, domain.ErrValidation("problem_id is required"))
		return
	}
	saved, err := h.repos.Practice.AddUpsolve(currentUsername(c), &item)
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, saved)
}

func (h *handlerImpl) practiceUpdateUpsolve(c *gin.Context) {
	if h.repos.Practice == nil {
		respondError(c, domain.ErrInternal("practice repository unavailable"))
		return
	}
	var req upsolveUpdateRequest
	if err := bindJSON(c, &req); err != nil {
		respondError(c, err)
		return
	}
	saved, err := h.repos.Practice.UpdateUpsolve(currentUsername(c), c.Param("problemId"), req.Resolved)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, saved)
}

func applyNewReviewDefaults(item *domain.ReviewItem) {
	if item.DueDate == "" {
		item.DueDate = time.Now().UTC().Format(time.RFC3339)
	}
	if item.Interval <= 0 {
		item.Interval = 1
	}
	if item.Ease == 0 {
		item.Ease = domain.DefaultReviewEase
	}
	if item.Ease < domain.MinReviewEase {
		item.Ease = domain.MinReviewEase
	}
	if item.Repetitions < 0 {
		item.Repetitions = 0
	}
}
