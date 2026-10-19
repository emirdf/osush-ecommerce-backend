package response

type SubcategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SubcategoryListResponse struct {
	Total         int                   `json:"total"`
	Subcategories []SubcategoryResponse `json:"subcategories"`
}
