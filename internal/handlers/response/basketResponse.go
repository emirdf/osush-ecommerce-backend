package response

type BasketResponse struct {
	Total int `json:"total"`
}

type BasketListResponse struct {
	Total      int             `json:"total"`
	TotalPrice float64         `json:"total_price"`
	Products   []ProductBasket `json:"products"`
}
