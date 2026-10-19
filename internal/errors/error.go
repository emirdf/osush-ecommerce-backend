package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrInternalServer   = NewAppError(nil, "internal server error", http.StatusInternalServerError)
	ErrUnauthorized     = NewAppError(nil, "authentication failed", http.StatusUnauthorized)
	ErrMissingToken     = NewAppError(nil, "no token", http.StatusUnauthorized)
	ErrTokenInBlacklist = NewAppError(nil, "token is blacklisted", http.StatusUnauthorized)
	ErrAlreadyDataa     = NewAppError(nil, "not found", http.StatusNotFound)
	ErrForbidden        = NewAppError(nil, "you don't have access", http.StatusForbidden)
	ErrBadToken         = NewAppError(nil, "invalid token", http.StatusUnauthorized)
)

var ErrNotFoundByID = func(id int, field string) *AppError {
	return NewAppError(
		nil,
		fmt.Sprintf("not found, %s id: %v", field, id),
		http.StatusNotFound,
	)
}

var ErrNotFound = func(field string) *AppError {
	return NewAppError(
		nil,
		fmt.Sprintf("not found, %s", field),
		http.StatusNotFound,
	)
}

var ErrAlreadyData = func(field string) *AppError {
	return NewAppError(
		nil,
		fmt.Sprintf("the %s already exists", field),
		http.StatusBadRequest,
	)
}

var ErrBadRequest = func(error string) *AppError {
	return NewAppError(
		nil,
		error,
		http.StatusBadRequest,
	)
}

type AppError struct {
	Err     error  `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(err error, message string, code int) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

func HandleError(c *gin.Context, err error) {
	var appErr *AppError

	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, appErr)
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   err.Error(),
		"message": ErrInternalServer.Message,
		"code":    http.StatusInternalServerError,
	})
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}

	return e.Message == t.Message &&
		e.Code == t.Code
}
