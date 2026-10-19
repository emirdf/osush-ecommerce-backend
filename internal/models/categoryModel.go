package models

import "time"

type Category struct {
	ID        int
	Name      string
	ImagePath string
	CreatedAt time.Time
	UpdatedAt time.Time
}
