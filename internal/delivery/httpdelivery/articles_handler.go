package httpdelivery

import (
	"strconv"
	"strings"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *handlerImpl) articlesList(c *gin.Context) {
	if h.repos.Article == nil {
		respondError(c, domain.ErrInternal("article repository unavailable"))
		return
	}
	filter := parseArticleFilter(c)
	articles, err := h.repos.Article.List(domain.ArticleFilter{
		Limit:  filter.Limit,
		Offset: filter.Offset,
		Source: filter.Source,
		Tag:    filter.Tag,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, articles)
}

func parseArticleFilter(c *gin.Context) domain.ArticleFilter {
	limit := 10
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := 0
	if raw := c.Query("offset"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			offset = n
		}
	}

	return domain.ArticleFilter{
		Limit:  limit,
		Offset: offset,
		Source: strings.TrimSpace(c.Query("source")),
		Tag:    strings.TrimSpace(c.Query("tag")),
	}
}
