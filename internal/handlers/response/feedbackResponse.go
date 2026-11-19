package response

type FeedbackClientResponse struct {
	ID          int     `json:"id"`
	UserName    string  `json:"user_name"`
	Surname     *string `json:"surname"`
	PhoneNumber string  `json:"phone_number"`
}

type FeedbackResponse struct {
	ID        int                     `json:"id"`
	Message   string                  `json:"message"`
	Client    *FeedbackClientResponse `json:"client"`
	CreatedAt string                  `json:"created_at"`
}

type FeedbackListResponse struct {
	Total     int                `json:"total"`
	Feedbacks []FeedbackResponse `json:"feedbacks"`
}
