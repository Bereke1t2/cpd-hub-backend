package httpdelivery

import (
	"github.com/gin-gonic/gin"
)

func (h *handlerImpl) learningTopics(c *gin.Context) {
	list, err := h.repos.Learning.GetTopics()
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, list)
}

func (h *handlerImpl) learningTracks(c *gin.Context) {
	list, err := h.repos.Learning.GetTracks()
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, list)
}

func (h *handlerImpl) learningLesson(c *gin.Context) {
	topicID := c.Param("topicId")
	l, err := h.repos.Learning.GetLesson(topicID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, l)
}
