package repository

// FigureListFilter 图表列表筛选（用户维度已在 repo 方法传入 userID）。
type FigureListFilter struct {
	Keyword string
	// CategoryID：nil=不限；指向 0=仅未分类；>0=指定分类
	CategoryID *uint64
	// ManuscriptID：nil=不限；指向 0=未关联任何论文；>0=指定论文
	ManuscriptID *uint64
}
