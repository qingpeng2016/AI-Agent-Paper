package request

type PaperFigureListQuery struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Q            string `form:"q"`
	CategoryID   string `form:"category_id"`   // 空=全部；none=未分类；否则为分类 id
	ManuscriptID string `form:"manuscript_id"` // 空=全部；none=未关联论文；否则为 paper_manuscript.id
}

type PaperFigureCategoryListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
