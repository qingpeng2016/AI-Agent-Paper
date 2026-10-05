package response

type UserNotificationListPageResp struct {
	Items    []UserNotificationItemResp `json:"items"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}

type UserNotificationItemResp struct {
	ID           uint   `json:"id"`
	Category     string `json:"category"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	TemplateCode string `json:"template_code"`
	Read         bool   `json:"read"`
	CreatedAt    string `json:"created_at"`
	LinkTab      string `json:"link_tab,omitempty"`
}

type UserNotificationUnreadCountResp struct {
	Count int64 `json:"count"`
}
