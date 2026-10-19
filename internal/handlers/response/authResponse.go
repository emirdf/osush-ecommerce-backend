package response

type UserProfileResponse struct {
	ID          int     `json:"id"`
	UserName    string  `json:"user_name"`
	Surname     *string `json:"surname"`
	PhoneNumber *string `json:"phone_number"`
	Avatar      *string `json:"avatar"`
}

type AdminIndexPageResponse struct {
	TotalClients    int                     `json:"total_clients"`
	TotalProducts   int                     `json:"total_products"`
	TotalCategories int                     `json:"total_categories"`
	Products        []AdminProductsResponse `json:"products"`
	Clients         []string                `json:"clients"`
}

type AdminProductsResponse struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	OrderCount    int      `json:"order_count"`
	Price         float64  `json:"price"`
	DiscountPrice *float64 `json:"discount_price"`
	Image         string   `json:"image"`
}
