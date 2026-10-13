package request

type PaperManuscriptListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
