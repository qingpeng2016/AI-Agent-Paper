// 官方参考：https://docs.openalex.org/
package entity

// OpenAlexSearchQuery GET /works?search=
type OpenAlexSearchQuery struct {
	Query      string
	MaxResults int
}

// OpenAlexWork 单条 OpenAlex work
type OpenAlexWork struct {
	ExternalKey     string
	Title           string
	Authors         []string
	PublishedYear   int
	DOI             string
	URL             string
}
