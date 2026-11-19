package request

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	FeedbackMessageMinLength = 3
	FeedbackMessageMaxLength = 1000
)

type FeedbackRequest struct {
	Message string `json:"message"`
}

func (r *FeedbackRequest) Validate() error {
	r.Message = strings.TrimSpace(r.Message)

	if r.Message == "" {
		return errors.New("message is required")
	}

	length := utf8.RuneCountInString(r.Message)

	if length < FeedbackMessageMinLength {
		return fmt.Errorf("message must be at least %d characters", FeedbackMessageMinLength)
	}

	if length > FeedbackMessageMaxLength {
		return fmt.Errorf("message must not exceed %d characters", FeedbackMessageMaxLength)
	}

	return nil
}
