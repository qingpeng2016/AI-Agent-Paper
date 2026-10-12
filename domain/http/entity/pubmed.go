// 官方参考：https://www.ncbi.nlm.nih.gov/books/NBK25501/
package entity

type PubMedSearchQuery struct {
	Query      string
	MaxResults int
	Offset     int
}

type PubMedArticle struct {
	ExternalKey   string
	Title         string
	Authors       []string
	Abstract      string
	PublishedYear int
	DOI           string
	URL           string
	PMCID         string // 如 PMC5159433（无 PMC 则为空）
}
