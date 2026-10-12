// 官方参考：https://docs.openalex.org/
package entity

// OpenAlexSearchQuery GET /works?search=
type OpenAlexSearchQuery struct {
	Query      string
	MaxResults int
	Page       int // 1-based，0 或未设等价于第 1 页
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
