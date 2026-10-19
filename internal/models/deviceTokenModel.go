package models

import "time"

type DeviceToken struct {
	ID          int
	ClientID    *int
	DeviceToken string
	UpdatedAt   time.Time
	CreatedAt   time.Time
}
