package response

type OrderResponse struct {
	ID              int                    `json:"id"`
	CustomerName    *string                `json:"customer_name"`
	PhoneNumber     string                 `json:"phone_number"`
	City            string                 `json:"city"`
	DeliveryAddress string                 `json:"delivery_address"`
	Notes           *string                `json:"notes"`
	TotalProduct    int                    `json:"total_product"`
	Products        []OrderProductResponse `json:"products"`
	TotalPrice      float64                `json:"total_price"`
	IsSend          bool                   `json:"is_send"`
}

type OrderProductResponse struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Price         float64  `json:"price"`
	DiscountPrice *float64 `json:"discount_price"`
	Image         string   `json:"image"`
	Total         int      `json:"total"`
}

type OrderProductListResponse struct {
	Total  int              `json:"total"`
	Orders []OrdersResponse `json:"orders"`
}

type OrdersResponse struct {
	ID           int     `json:"id"`
	CustomerName *string `json:"customer_name"`
	PhoneNumber  string  `json:"phone_number"`
	City         string  `json:"city"`
	TotalProduct int     `json:"total_product"`
	TotalPrice   float64 `json:"total_price"`
	IsSend       bool    `json:"is_send"`
	IsNew        bool    `json:"is_new"`
}

type OrderProductClientListResponse struct {
	Total  int                    `json:"total"`
	Orders []OrdersClientResponse `json:"orders"`
}

type OrdersClientResponse struct {
	ID              int     `json:"id"`
	DeliveryAddress string  `json:"delivery_address"`
	TotalProduct    int     `json:"total_product"`
	TotalPrice      float64 `json:"total_price"`
	IsSend          bool    `json:"is_send"`
}
