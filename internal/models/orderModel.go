package models

import "time"

type Order struct {
	ID              int
	ClientID        int
	CustomerName    *string
	PhoneNumber     string
	City            string
	DeliveryAddress string
	Notes           *string
	IsNew           bool
	IsSend          bool
	UpdatedAt       time.Time
	CreatedAt       time.Time
}
