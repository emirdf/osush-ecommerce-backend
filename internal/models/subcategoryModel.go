package models

import "time"

type Subcategory struct {
	ID         int
	CategoryID int
	Name     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
