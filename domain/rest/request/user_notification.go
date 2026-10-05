package request

type ListUserNotificationsQuery struct {
	Page        int  `form:"page"`
	PageSize    int  `form:"page_size"`
	UnreadOnly  bool `form:"unread_only"`
}
