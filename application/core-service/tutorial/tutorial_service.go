package tutorial

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/shared"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

var ErrTutorialNotFound = errors.New("tutorial not found")

const tutorialListPageTitle = "AI Plan 教程与资讯"

type TutorialService struct {
	categoryRepo repository.TutorialCategoryRepo
	articleRepo  repository.TutorialArticleRepo
}

func NewTutorialService(categoryRepo repository.TutorialCategoryRepo, articleRepo repository.TutorialArticleRepo) *TutorialService {
	return &TutorialService{categoryRepo: categoryRepo, articleRepo: articleRepo}
}

func (s *TutorialService) List(ctx context.Context, categoryCode string) (*response.TutorialListResp, error) {
	categories, err := s.categoryRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	articles, err := s.articleRepo.ListPublished(ctx, categoryCode)
	if err != nil {
		return nil, err
	}

	catByID := map[uint]entity.TutorialCategory{}
	catResp := make([]response.TutorialCategoryResp, 0, len(categories))
	for i := range categories {
		c := categories[i]
		catByID[c.ID] = c
		catResp = append(catResp, response.TutorialCategoryResp{
			ID:   c.ID,
			Code: c.Code,
			Name: c.Name,
			Sort: c.Sort,
		})
	}

	items := make([]response.TutorialArticleItemResp, 0, len(articles))
	for i := range articles {
		a := &articles[i]
		items = append(items, s.mapArticleItem(a, catByID))
	}

	return &response.TutorialListResp{
		Title:      tutorialListPageTitle,
		Categories: catResp,
		Articles:   items,
	}, nil
}

func (s *TutorialService) Detail(ctx context.Context, slug string) (*response.TutorialArticleItemResp, error) {
	article, err := s.articleRepo.FindPublishedBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, ErrTutorialNotFound
	}
	cat, err := s.categoryRepo.GetByID(ctx, article.CategoryID)
	if err != nil {
		return nil, err
	}
	catByID := map[uint]entity.TutorialCategory{}
	if cat != nil {
		catByID[cat.ID] = *cat
	}
	item := s.mapArticleItem(article, catByID)
	item.Body = shared.DecodeStringJSONArray(article.Body)
	return &item, nil
}

func (s *TutorialService) mapArticleItem(
	a *entity.TutorialArticle,
	catByID map[uint]entity.TutorialCategory,
) response.TutorialArticleItemResp {
	code := ""
	name := ""
	if c, ok := catByID[a.CategoryID]; ok {
		code = c.Code
		name = c.Name
	}
	return response.TutorialArticleItemResp{
		Slug:         a.Slug,
		CategoryCode: code,
		CategoryName: name,
		Date:         a.PublishedAt.Format("2006-01-02"),
		Title:        a.Title,
		Excerpt:      a.Excerpt,
	}
}
