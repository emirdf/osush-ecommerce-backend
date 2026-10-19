package response

type NotificationResponse struct {
	ID        int        `json:"id"`
	Title     Dictionary `json:"title"`
	Body      Dictionary `json:"body"`
	IsNew     bool       `json:"is_new"`
	CreatedAt string     `json:"created_at"`
}

type NotificationListResponse struct {
	TotalNotifications int                    `json:"total_notifications"`
	TotalNew           int                    `json:"total_new"`
	Notifications      []NotificationResponse `json:"notifications"`
}

type NotificationResponseForAdmin struct {
	ID        int        `json:"id"`
	Title     Dictionary `json:"title"`
	Body      Dictionary `json:"body"`
	Life      int        `json:"life"`
	CreatedAt string     `json:"created_at"`
}

type NotificationListResponseForAdmin struct {
	Total         int                            `json:"total"`
	Notifications []NotificationResponseForAdmin `json:"notifications"`
}
