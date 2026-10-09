package request

type PaperManuscriptSetCurrentRequest struct {
	ManuscriptID uint64 `json:"manuscript_id" binding:"required"`
}

type PaperManuscriptCreateRequest struct {
	Title          string `json:"title" binding:"required"`
	DisciplineCode string `json:"discipline_code" binding:"required"`
}
