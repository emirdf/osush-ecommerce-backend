package request

import (
	"errors"
	"osush/internal/validation"
)

type RegisterRequest struct {
	UserName    string  `json:"user_name" binding:"required"`
	Surname     *string `json:"surname"`
	Address     *string `json:"address"`
	Password    string  `json:"password" binding:"required"`
	PhoneNumber string  `json:"phone_number" binding:"required"`
}

func (r *RegisterRequest) Validate() error {
	if r.UserName == "" {
		return errors.New("user_name is required")
	}

	if err := validation.ValidateUsername(r.UserName, "user_name"); err != nil {
		return err
	}

	if r.Surname != nil {
		if err := validation.ValidateTextSize(*r.Surname, "surname"); err != nil {
			return err
		}
	}

	if r.Password == "" {
		return errors.New("password is required")
	}

	if r.Address != nil {
		if err := validation.ValidateAddress(*r.Address); err != nil {
			return err
		}
	}

	if r.PhoneNumber == "" {
		return errors.New("phone_number is required")
	}

	phoneNumber, err := validation.ValidatePhone(r.PhoneNumber)
	if err != nil {
		return err
	}

	r.PhoneNumber = phoneNumber

	return nil
}

type LoginClientRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

func (r *LoginClientRequest) Validate() error {
	if r.PhoneNumber == "" {
		return errors.New("phone_number is required")
	}

	phoneNumber, err := validation.ValidatePhone(r.PhoneNumber)
	if err != nil {
		return err
	}

	r.PhoneNumber = phoneNumber

	if r.Password == "" {
		return errors.New("password is required")
	}

	if err := validation.ValidatePassword(r.Password); err != nil {
		return err
	}

	return nil
}

type UpdateRequest struct {
	UserName    string  `json:"user_name" binding:"required"`
	Surname     *string `json:"surname"`
	Address     *string `json:"address"`
	PhoneNumber string  `json:"phone_number" binding:"required"`
}

func (r *UpdateRequest) Validate() error {
	if r.UserName == "" {
		return errors.New("user_name is required")
	}

	if err := validation.ValidateUsername(r.UserName, "user_name"); err != nil {
		return err
	}

	if r.Surname != nil {
		if err := validation.ValidateTextSize(*r.Surname, "surname"); err != nil {
			return err
		}
	}

	if r.Address != nil {
		if err := validation.ValidateAddress(*r.Address); err != nil {
			return err
		}
	}

	if r.PhoneNumber == "" {
		return errors.New("phone_number is required")
	}

	phoneNumber, err := validation.ValidatePhone(r.PhoneNumber)
	if err != nil {
		return err
	}

	r.PhoneNumber = phoneNumber

	return nil
}
