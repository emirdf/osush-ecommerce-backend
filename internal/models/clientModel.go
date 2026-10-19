package models

import "time"

type Client struct {
	ID          int
	UserName    string
	Surname     *string
	PhoneNumber string
	Address     *string
	ImagePath   *string
	Password    string
	UpdatedAt   time.Time
	CreatedAt   time.Time
}
