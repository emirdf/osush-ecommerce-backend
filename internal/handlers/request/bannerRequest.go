package request

import (
	"errors"
	"osush/internal/validation"
)

type BannerRequest struct {
	Title       Dictionary `json:"title"`
	URL         *string    `json:"url"`
	Description Dictionary `json:"description"`
	PhoneNumber *string    `json:"phone_number"`
	DisplayFrom string     `json:"display_from"`
	DisplayTo   string     `json:"display_to"`
}

func (r *BannerRequest) Validate() error {
	if r.Title.Tm == "" {
		return errors.New("title tm is required")
	}

	if r.Title.Ru == "" {
		return errors.New("title ru is required")
	}

	if r.Description.Tm != "" && r.Description.Ru != "" {
		if err := validation.ValidateTextSize(
			r.Description.Tm,
			"description in tm",
		); err != nil {
			return err
		}

		if err := validation.ValidateTextSize(
			r.Description.Ru,
			"description in ru",
		); err != nil {
			return err
		}
	} else if r.Description.Tm != "" {
		return errors.New("description ru is required")
	} else if r.Description.Ru != "" {
		return errors.New("description tm is required")
	}

	if r.PhoneNumber != nil {
		phoneNumber, err := validation.ValidatePhone(*r.PhoneNumber)
		if err != nil {
			return err
		}

		r.PhoneNumber = &phoneNumber
	}

	if r.DisplayFrom == "" {
		return errors.New("display_from is required")
	}

	if err := validation.ValidateDate(r.DisplayFrom); err != nil {
		return err
	}

	if r.DisplayTo == "" {
		return errors.New("display_to is required")
	}

	if err := validation.ValidateDate(r.DisplayTo); err != nil {
		return err
	}

	if err := validation.ValidateDateRange(r.DisplayFrom, r.DisplayTo); err != nil {
		return err
	}

	return nil
}
