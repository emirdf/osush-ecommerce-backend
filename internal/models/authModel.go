package models

import "time"

type Users struct {
	ID          int
	Name        string
	Surname     *string
	PhoneNumber *string
	Avatar      *string
	Password    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
