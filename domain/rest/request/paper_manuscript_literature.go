package request

type PaperManuscriptLiteratureIDsQuery struct {
	ManuscriptID string `form:"manuscript_id"`
}

type PaperManuscriptSetLiteratureMembershipRequest struct {
	ManuscriptID  string   `json:"manuscript_id"`
	LiteratureIDs []string `json:"literature_ids"`
}

type PaperManuscriptWithLiteratureListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
