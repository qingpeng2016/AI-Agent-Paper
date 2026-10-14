package response

type PaperManuscriptSubmissionListView struct {
	ManuscriptID uint64                              `json:"manuscript_id"`
	Items        []PaperManuscriptSubmissionItemView `json:"items"`
	Total        int64                               `json:"total,omitempty"`
	Page         int                                 `json:"page,omitempty"`
	PageSize     int                                 `json:"page_size,omitempty"`
}

type PaperManuscriptSubmissionItemView struct {
	ID              string `json:"id"`
	VenueCode       string `json:"venue_code"`
	VenueLabel      string `json:"venue_label"`
	Track           string `json:"track"`
	Round           int    `json:"round"`
	Status          string `json:"status"`
	DraftVersion    int    `json:"draft_version"`
	Anonymous       bool   `json:"anonymous"`
	PortalURL       string `json:"portal_url,omitempty"`
	ExternalID      string `json:"external_id,omitempty"`
	Title           string `json:"title,omitempty"`
	Abstract        string `json:"abstract,omitempty"`
	MainPDFFilename string `json:"main_pdf_filename,omitempty"`
	Decision        string `json:"decision,omitempty"`
	CommentsSummary string `json:"comments_summary,omitempty"`
	RevisionDeadline string `json:"revision_deadline,omitempty"`
	SubmittedAt     string `json:"submitted_at,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type PaperManuscriptSubmissionDetailView struct {
	PaperManuscriptSubmissionItemView
	Checklist []PaperManuscriptSubmissionChecklistItemView `json:"checklist,omitempty"`
	Events    []PaperManuscriptSubmissionEventView         `json:"events,omitempty"`
}

type PaperManuscriptSubmissionChecklistItemView struct {
	ID  string `json:"id"`
	OK  bool   `json:"ok"`
	Label string `json:"label,omitempty"`
}

type PaperManuscriptSubmissionEventView struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}
