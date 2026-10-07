// 官方参考：https://api.semanticscholar.org/api-docs/graph
package entity

// SemanticScholarSearchQuery GET /graph/v1/paper/search
type SemanticScholarSearchQuery struct {
	Query      string
	MaxResults int
}

// SemanticScholarPaper 单条 Semantic Scholar paper
type SemanticScholarPaper struct {
	ExternalKey     string
	Title           string
	Authors         []string
	Abstract        string
	PublishedYear   int
	DOI             string
	URL             string
}
