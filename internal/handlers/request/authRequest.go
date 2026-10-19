package request

import (
	"errors"
	"osush/internal/validation"
)

type LoginRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	if r.UserName == "" {
		return errors.New("user_name is required")
	}

	if err := validation.ValidateUsername(r.UserName, "user_name"); err != nil {
		return err
	}

	if r.Password == "" {
		return errors.New("password is required")
	}

	if err := validation.ValidatePassword(r.Password); err != nil {
		return err
	}

	return nil
}

type UpdateProfileRequest struct {
	Name            string `json:"user_name"`
	Surname         string `json:"surname"`
	PhoneNumber     string `json:"phone_number"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (r *UpdateProfileRequest) Validate() error {

	if r.Name == "" {
		return errors.New("user_name is required")
	}

	if (r.CurrentPassword != "" && r.NewPassword == "") ||
		(r.CurrentPassword == "" && r.NewPassword != "") {

		return errors.New(
			"current password and new password must be provided together",
		)
	}

	if r.NewPassword != "" {
		if err := validation.ValidatePassword(r.NewPassword); err != nil {
			return err
		}
	}

	return nil
}
