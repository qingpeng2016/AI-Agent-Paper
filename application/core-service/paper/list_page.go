package paper

const DefaultListPageSize = 20

func NormalizeListPage(page, pageSize int) (normPage, normPageSize, offset int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultListPageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize, (page - 1) * pageSize
}
