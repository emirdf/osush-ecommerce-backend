package models

import "time"

type OrderProduct struct {
	ID               int
	OrderID          int
	ProductID        int
	Quantity         int
	Price            float64
	DiscountPrice    *float64
	ProductName      string
	ProductImagePath string
	UpdatedAt        time.Time
	CreatedAt        time.Time
}
