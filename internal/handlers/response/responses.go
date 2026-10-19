package response

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type Dictionary struct {
	Tm string `json:"tm"`
	Ru string `json:"ru"`
}
