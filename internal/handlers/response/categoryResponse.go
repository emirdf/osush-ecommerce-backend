package response

type CategoryResponse struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	ImagePath string `json:"image_path"`
}

type CategoryListResponse struct {
	Total      int                `json:"total"`
	Categories []CategoryResponse `json:"categories"`
}
