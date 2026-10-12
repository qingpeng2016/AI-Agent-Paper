// 官方参考：https://api.crossref.org/swagger-ui/index.html
package entity

type CrossrefSearchQuery struct {
	Query      string
	MaxResults int
	Offset     int
}

type CrossrefWork struct {
	ExternalKey   string
	Title         string
	Authors       []string
	PublishedYear int
	DOI           string
	URL           string
}
