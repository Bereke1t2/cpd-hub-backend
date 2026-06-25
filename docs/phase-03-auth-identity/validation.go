//go:build ignore
// Template for Phase 3 — copy to: internal/delivery/httpdelivery/validation.go
//
// Turns gin/validator binding errors into a single clean message, mapped to a
// domain validation error (400). Use in auth handlers and any POST/PUT body.
package httpdelivery

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// bindJSON binds + validates the request body, returning a *domain.AppError on failure.
func bindJSON(c *gin.Context, dst interface{}) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		return domain.ErrValidation(humanizeValidation(err))
	}
	return nil
}

func humanizeValidation(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msgs := make([]string, 0, len(ve))
		for _, fe := range ve {
			msgs = append(msgs, fieldMessage(fe))
		}
		return strings.Join(msgs, "; ")
	}
	// not a validator error (e.g. malformed JSON)
	return "invalid request body"
}

func fieldMessage(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", field, fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s must be alphanumeric", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
