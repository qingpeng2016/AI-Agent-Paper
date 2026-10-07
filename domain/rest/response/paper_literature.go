package response

type PaperLiteratureHit struct {
	SourceCode    string   `json:"source_code"`
	ExternalKey   string   `json:"external_key"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors,omitempty"`
	Abstract      string   `json:"abstract,omitempty"`
	PublishedYear int      `json:"published_year,omitempty"`
	DOI           string   `json:"doi,omitempty"`
	URL           string   `json:"url,omitempty"`
}

type PaperLiteratureSearchResult struct {
	Query      string               `json:"query"`
	SourceCode string               `json:"source_code"`
	Hits       []PaperLiteratureHit `json:"hits"`
}
