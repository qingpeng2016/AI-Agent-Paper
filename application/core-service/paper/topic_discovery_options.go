package paper

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

type TopicDiscoveryOptionsService struct {
	catalog   repository.PaperRefCatalogRepo
	litSource repository.PaperLiteraturePlatformRepo
}

func NewTopicDiscoveryOptionsService(
	catalog repository.PaperRefCatalogRepo,
	litSource repository.PaperLiteraturePlatformRepo,
) *TopicDiscoveryOptionsService {
	return &TopicDiscoveryOptionsService{catalog: catalog, litSource: litSource}
}

func (s *TopicDiscoveryOptionsService) GetFormOptions(ctx context.Context) (*response.TopicDiscoveryFormOptions, error) {
	disciplines, err := s.catalog.ListActiveDisciplines(ctx)
	if err != nil {
		return nil, err
	}
	intensities, err := s.catalog.ListExecutionIntensities(ctx)
	if err != nil {
		return nil, err
	}
	audits, err := s.catalog.ListAuditLevels(ctx, entity.LLMSceneTopic)
	if err != nil {
		return nil, err
	}
	venues, err := s.catalog.ListActiveVenues(ctx)
	if err != nil {
		return nil, err
	}

	out := &response.TopicDiscoveryFormOptions{
		Disciplines:      make([]response.TopicDiscoveryDisciplineOption, 0, len(disciplines)),
		Venues:           make([]response.TopicDiscoveryVenueOption, 0, len(venues)),
		ExecutionIntents: make([]response.TopicDiscoveryExecutionIntensityOption, 0, len(intensities)),
		AuditLevels:      make([]response.TopicDiscoveryAuditLevelOption, 0, len(audits)),
	}

	for _, v := range venues {
		out.Venues = append(out.Venues, response.TopicDiscoveryVenueOption{
			Code:      v.Code,
			Label:     v.Name,
			VenueType: v.VenueType,
			Sort:      v.Sort,
		})
	}

	for _, d := range disciplines {
		opt := response.TopicDiscoveryDisciplineOption{
			Code:  d.Code,
			Label: d.Name,
			Sort:  d.Sort,
		}
		if d.NameEn != nil {
			opt.NameEn = *d.NameEn
		}
		out.Disciplines = append(out.Disciplines, opt)
	}

	for _, i := range intensities {
		out.ExecutionIntents = append(out.ExecutionIntents, response.TopicDiscoveryExecutionIntensityOption{
			Code:       i.Code,
			Label:      i.Name,
			ExecRounds: i.ExecRounds,
			MaxPapers:  i.MaxPapers,
			MaxIdeas:   i.MaxIdeas,
			IsDefault:  i.IsDefault,
		})
	}

	for _, a := range audits {
		out.AuditLevels = append(out.AuditLevels, response.TopicDiscoveryAuditLevelOption{
			Code:            a.Code,
			Label:           a.Name,
			LlmDescriptionCitationCheckCode: a.LlmDescriptionCitationCheckCode,
			LlmDescriptionClaimCheckCode:    a.LlmDescriptionClaimCheckCode,
			LlmDescriptionKillCheckCode:     a.LlmDescriptionKillCheckCode,
			IsDefault:       a.IsDefault,
		})
	}

	sources, err := s.litSource.ListActiveOrdered(ctx)
	if err != nil {
		return nil, err
	}
	out.LiteratureSources = make([]response.TopicDiscoveryLiteratureSourceOption, 0, len(sources))
	for _, src := range sources {
		out.LiteratureSources = append(out.LiteratureSources, response.TopicDiscoveryLiteratureSourceOption{
			Code:            src.Code,
			Label:           src.Name,
			Sort:            src.Sort,
			DefaultSelected: src.DefaultSelected,
		})
	}

	return out, nil
}
