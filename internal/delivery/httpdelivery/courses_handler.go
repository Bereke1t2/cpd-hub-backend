package httpdelivery

import (
	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *handlerImpl) coursesList(c *gin.Context) {
	if h.repos.Course == nil {
		respondError(c, domain.ErrInternal("course repository unavailable"))
		return
	}
	courses, err := h.repos.Course.List(currentUsername(c))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, courses)
}

func (h *handlerImpl) coursesGet(c *gin.Context) {
	if h.repos.Course == nil {
		respondError(c, domain.ErrInternal("course repository unavailable"))
		return
	}
	course, err := h.repos.Course.Get(currentUsername(c), c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, course)
}

func (h *handlerImpl) coursesCompleteLesson(c *gin.Context) {
	if h.repos.Course == nil {
		respondError(c, domain.ErrInternal("course repository unavailable"))
		return
	}
	lessonID := c.Param("lessonId")
	if err := h.repos.Course.CompleteLesson(currentUsername(c), c.Param("courseId"), lessonID); err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, domain.LessonCompletion{LessonID: lessonID, Completed: true})
}
