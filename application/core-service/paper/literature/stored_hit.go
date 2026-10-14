package literature

// StoredLiteratureHit 选题 retrieve 落库的单条文献命中（step.result.literature_hits）。
type StoredLiteratureHit struct {
	SourceID       uint     `json:"source_id"`
	SourceCode     string   `json:"source_code"`
	ExternalKey    string   `json:"external_key"`
	Title          string   `json:"title"`
	Authors        []string `json:"authors,omitempty"`
	URL            string   `json:"url,omitempty"`
	DOI            string   `json:"doi,omitempty"`
	PublishedYear  int      `json:"published_year,omitempty"`
	RelevanceScore float64  `json:"relevance_score,omitempty"`
	QueryText      string   `json:"query_text,omitempty"`
	Meta           any      `json:"meta,omitempty"`
}
