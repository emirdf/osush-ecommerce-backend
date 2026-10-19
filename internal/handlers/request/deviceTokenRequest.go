package request

import (
	"errors"
)

type DeviceTokenRequest struct {
	DeviceToken string `json:"device_token"`
}

func (r *DeviceTokenRequest) Validate() error {
	if r.DeviceToken == "" {
		return errors.New("device_token is required")
	}
	return nil
}
