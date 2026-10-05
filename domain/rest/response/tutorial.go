package response

type TutorialListResp struct {
	Title      string                  `json:"title"`
	Categories []TutorialCategoryResp  `json:"categories"`
	Articles   []TutorialArticleItemResp `json:"articles"`
}

type TutorialCategoryResp struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

type TutorialArticleItemResp struct {
	Slug         string   `json:"slug"`
	CategoryCode string   `json:"category_code"`
	CategoryName string   `json:"category_name"`
	Date         string   `json:"date"`
	Title        string   `json:"title"`
	Excerpt      string   `json:"excerpt"`
	Body         []string `json:"body,omitempty"`
}
