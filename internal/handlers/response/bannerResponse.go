package response

type BannerResponse struct {
	ID          int         `json:"id"`
	Title       Dictionary  `json:"title"`
	URL         *string     `json:"url"`
	Description *Dictionary `json:"description"`
	PhoneNumber *string     `json:"phone_number"`
	Image       string      `json:"image"`
	Images      []string    `json:"images"`
	DisplayFrom string      `json:"display_from"`
	DisplayTo   string      `json:"display_to"`
	IsActive    bool        `json:"is_active"`
}

type BannersResponse struct {
	Total   int              `json:"total"`
	Banners []BannerResponse `json:"banners"`
}
