package httpdelivery

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// bindJSON binds + validates the request body.
// Returns a *domain.AppError (validation, 400) on failure.
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
