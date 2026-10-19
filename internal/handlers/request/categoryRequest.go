package request

import (
	"errors"
)

type CategoryRequest struct {
	Name string `json:"name"`
}

func (r *CategoryRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is tm required")
	}
	return nil
}

