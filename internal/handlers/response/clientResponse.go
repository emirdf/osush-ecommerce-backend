package response

type ProfileResponse struct {
	ID          int     `json:"id"`
	UserName    string  `json:"user_name"`
	Surname     *string `json:"surname"`
	PhoneNumber string  `json:"phone_number"`
	Address     *string `json:"address"`
	ImagePath   *string `json:"avatar"`
}

type ClientListResponse struct {
	Total   int      `json:"total"`
	Clients []Client `json:"clients"`
}

type Client struct {
	ID          int     `json:"id"`
	UserName    string  `json:"user_name"`
	Surname     *string `json:"surname"`
	PhoneNumber string  `json:"phone_number"`
	OrderCount  int     `json:"order_count"`
	TotalSpent  float64 `json:"total_spent"`
}
