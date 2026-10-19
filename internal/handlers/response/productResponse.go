package response

import (
	"osush/internal/handlers/request"
)

type ProductResponse struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	IsLiked       bool                   `json:"is_liked"`
	Description   *Dictionary            `json:"description"`
	Price         float64                `json:"price"`
	DiscountPrice *float64               `json:"discount_price"`
	Attributes    *request.DictionaryMap `json:"attributes"`
	IsPaused      bool                   `json:"is_paused"`
	Images        []string               `json:"images"`
}

type ProductListResponse struct {
	Total    int        `json:"total"`
	Products []Products `json:"products"`
}

type Products struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	IsLiked       bool     `json:"is_liked"`
	Price         float64  `json:"price"`
	DiscountPrice *float64 `json:"discount_price"`
	Image         string   `json:"image"`
	IsPaused      bool     `json:"is_paused"`
	OrderCount    int      `json:"order_count"`
	CreatedAt     string   `json:"created_at"`
}

type IndexProductResponse struct {
	Products    []ProductIndex `json:"products"`
	NewProducts []ProductIndex `json:"new_products"`
}

type ProductIndex struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	IsLiked       bool     `json:"is_liked"`
	Price         float64  `json:"price"`
	DiscountPrice *float64 `json:"discount_price"`
	Image         string   `json:"image"`
	IsPaused      bool     `json:"is_paused"`
}

type ProductBasket struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Price         float64  `json:"price"`
	DiscountPrice *float64 `json:"discount_price"`
	Image         string   `json:"image"`
	IsPaused      bool     `json:"is_paused"`
	Total         int      `json:"total"`
}
