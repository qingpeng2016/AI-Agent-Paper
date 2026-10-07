// 官方参考：https://info.arxiv.org/help/api/index.html
package entity

// ArxivSearchQuery export.arxiv.org/api/query
type ArxivSearchQuery struct {
	Query      string
	MaxResults int
}

// ArxivPaper 单条 arXiv 条目（Atom entry 映射）
type ArxivPaper struct {
	ExternalKey   string
	Title         string
	Authors       []string
	Abstract      string
	PublishedYear int
	URL           string
}
