package request

type PaperManuscriptSubmissionListQuery struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

type PaperManuscriptSubmissionDetailQuery struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
	SubmissionID uint64 `form:"submission_id" binding:"required"`
}

type PaperManuscriptSubmissionCreateBody struct {
	ManuscriptID    uint64 `json:"manuscript_id" binding:"required"`
	VenueCode       string `json:"venue_code" binding:"required"`
	VenueLabel      string `json:"venue_label" binding:"required"`
	Track           string `json:"track"`
	DraftVersion    int    `json:"draft_version" binding:"required"`
	Anonymous       bool   `json:"anonymous"`
	PortalURL       string `json:"portal_url"`
	ExternalID      string `json:"external_id"`
	Title           string `json:"title"`
	Abstract        string `json:"abstract"`
	MarkedSubmitted bool   `json:"marked_submitted"`
}

type PaperManuscriptSubmissionRevisionBody struct {
	ManuscriptID uint64 `json:"manuscript_id" binding:"required"`
	SubmissionID uint64 `json:"submission_id" binding:"required"`
	Decision     string `json:"decision"`
	Comments     string `json:"comments"`
}

type PaperManuscriptSubmissionMarkResubmittedBody struct {
	ManuscriptID uint64 `json:"manuscript_id" binding:"required"`
	SubmissionID uint64 `json:"submission_id" binding:"required"`
}

type PaperManuscriptSubmissionUploadForm struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
	SubmissionID uint64 `form:"submission_id" binding:"required"`
}
