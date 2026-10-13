package request

type PaperUserLiteratureListQuery struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Q          string `form:"q"`
	CategoryID string `form:"category_id"` // 空=全部；none=未分类；否则为分类 id
}

type PaperUserLiteratureCategoryListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
