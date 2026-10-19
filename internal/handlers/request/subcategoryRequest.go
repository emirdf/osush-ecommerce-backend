package request

import (
	"errors"
)

type SubcategoryRequest struct {
	CategoryID int    `json:"category_id"`
	Name       string `json:"name"`
}

func (r *SubcategoryRequest) Validate() error {
	if r.CategoryID == 0 {
		return errors.New("category_id is not zero, it is required")
	}

	if r.Name == "" {
		return errors.New("name is tm required")
	}
	return nil
}

type SubcategoryUpdateRequest struct {
	Name string `json:"name"`
}

func (r *SubcategoryUpdateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is tm required")
	}
	return nil
}
