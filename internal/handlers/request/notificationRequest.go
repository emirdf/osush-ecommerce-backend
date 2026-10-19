package request

import "errors"

type NotificationRequest struct {
	Title Dictionary `json:"title"`
	Body  Dictionary `json:"body"`
	Life  int        `json:"life"`
}

func (r *NotificationRequest) Validate() error {
	if r.Title.Tm == "" {
		return errors.New("title in tm is required")
	}

	if r.Title.Ru == "" {
		return errors.New("title in ru is required")
	}

	if r.Body.Tm == "" {
		return errors.New("body in tm is required")
	}

	if r.Body.Ru == "" {
		return errors.New("body in ru is required")
	}

	if r.Life <= 0 {
		return errors.New("life must be greater than 0")
	}

	return nil
}
