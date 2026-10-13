package response

type PaperUserLiteratureItemView struct {
	ID              string `json:"id"`
	From            string `json:"from"`
	FileExt         string `json:"file_ext"`
	FileName        string `json:"file_name"`
	LiteratureTitle string `json:"literature_title"`
	CategoryIDs      []string `json:"category_ids,omitempty"`
	CategoryNames    []string `json:"category_names,omitempty"`
	ManuscriptIDs    []string `json:"manuscript_ids,omitempty"`
	ManuscriptTitles []string `json:"manuscript_titles,omitempty"`
	Note            string `json:"note,omitempty"`
	SizeBytes       uint64 `json:"size_bytes"`
	StorageURI      string `json:"storage_uri,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

type PaperUserLiteratureListView struct {
	Items    []PaperUserLiteratureItemView `json:"items"`
	Total    int64                         `json:"total,omitempty"`
	Page     int                           `json:"page,omitempty"`
	PageSize int                           `json:"page_size,omitempty"`
}
