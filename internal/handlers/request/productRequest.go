package request

import (
	"errors"
	"osush/internal/validation"
)

type ProductRequest struct {
	Name          string        `json:"name" binding:"required"`
	SubcategoryID int           `json:"subcategory_id" binding:"required"`
	Description   Dictionary    `json:"description"`
	Price         float64       `json:"price" binding:"required"`
	DiscountPrice *float64      `json:"discount_price"`
	Attributes    DictionaryMap `json:"attributes"`
	IsPaused      bool          `json:"is_paused"`
}

func (r *ProductRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if r.SubcategoryID <= 0 {
		return errors.New("the name must not be empty")
	}

	if err := validation.ValidateTextSize(r.Description.Tm, "description in tm"); err != nil {
		return err
	}

	if err := validation.ValidateTextSize(r.Description.Ru, "description in ru"); err != nil {
		return err
	}

	if r.Price == 0 {
		return errors.New("the price must not be empty")
	}

	if r.DiscountPrice != nil && *r.DiscountPrice > r.Price {
		return errors.New("discount price must not be greater than the regular price")
	}

	if err := validation.ValidateMapJSONSize(r.Attributes.Tm, "Attributes in tm"); err != nil {
		return err
	}

	if err := validation.ValidateMapJSONSize(r.Attributes.Ru, "Attributes in ru"); err != nil {
		return err
	}

	return nil
}

type ProductFilter struct {
	SortByValue   string `form:"sort_by_value" binding:"omitempty,oneof=ASC DESC"`
	Search        string `form:"search"`
	Limit         string `form:"limit"`
	Offset        string `form:"offset"`
	SubcategoryID int    `form:"subcategory_id"`
	CategoryID    int    `form:"category_id"`
	LimitInt      int
	OffsetInt     int
}

type UpdateProductRequest struct {
	Name          string        `json:"name" binding:"required"`
	Description   Dictionary    `json:"description"`
	Price         float64       `json:"price" binding:"required"`
	DiscountPrice *float64      `json:"discount_price"`
	Attributes    DictionaryMap `json:"attributes"`
	IsPaused      bool          `json:"is_paused"`
}

func (r *UpdateProductRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}

	if err := validation.ValidateTextSize(r.Description.Tm, "description in tm"); err != nil {
		return err
	}

	if err := validation.ValidateTextSize(r.Description.Ru, "description in ru"); err != nil {
		return err
	}

	if r.Price == 0 {
		return errors.New("the price must not be empty")
	}

	if r.DiscountPrice != nil && *r.DiscountPrice > r.Price {
		return errors.New("discount price must not be greater than the regular price")
	}

	if err := validation.ValidateMapJSONSize(r.Attributes.Tm, "Attributes in tm"); err != nil {
		return err
	}

	if err := validation.ValidateMapJSONSize(r.Attributes.Ru, "Attributes in ru"); err != nil {
		return err
	}

	return nil
}
