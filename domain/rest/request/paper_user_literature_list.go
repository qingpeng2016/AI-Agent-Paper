package request

type PaperUserLiteratureListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
