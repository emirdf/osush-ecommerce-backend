package models

import "time"

type Basket struct {
	ID        int
	ClientID  int
	ProductID int
	Total     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BasketProduct struct {
	ID            int
	Name          string
	Price         float64
	DiscountPrice *float64
	Image         string
	IsPaused      bool
	Total         int
}
