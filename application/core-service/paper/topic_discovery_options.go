package paper

import (
	"context"
	"encoding/json"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

type TopicDiscoveryOptionsService struct {
	catalog repository.PaperRefCatalogRepo
}

func NewTopicDiscoveryOptionsService(catalog repository.PaperRefCatalogRepo) *TopicDiscoveryOptionsService {
	return &TopicDiscoveryOptionsService{catalog: catalog}
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
	audits, err := s.catalog.ListAuditLevels(ctx)
	if err != nil {
		return nil, err
	}

	out := &response.TopicDiscoveryFormOptions{
		Disciplines:      make([]response.TopicDiscoveryDisciplineOption, 0, len(disciplines)),
		ExecutionIntents: make([]response.TopicDiscoveryExecutionIntensityOption, 0, len(intensities)),
		AuditLevels:      make([]response.TopicDiscoveryAuditLevelOption, 0, len(audits)),
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
		if len(d.LiteratureSourceCodes) > 0 {
			var codes []string
			if err := json.Unmarshal(d.LiteratureSourceCodes, &codes); err == nil {
				opt.Sources = codes
			}
		}
		out.Disciplines = append(out.Disciplines, opt)
	}

	for _, i := range intensities {
		mul, _ := i.Multiplier.Float64()
		out.ExecutionIntents = append(out.ExecutionIntents, response.TopicDiscoveryExecutionIntensityOption{
			Code:       i.Code,
			Label:      i.Name,
			Multiplier: mul,
			MaxPapers:  i.MaxPapers,
			MaxIdeas:   i.MaxIdeas,
			IsDefault:  i.IsDefault,
		})
	}

	for _, a := range audits {
		out.AuditLevels = append(out.AuditLevels, response.TopicDiscoveryAuditLevelOption{
			Code:                 a.Code,
			Label:                a.Name,
			CitationStrength:     a.CitationStrength,
			ClaimStrength:        a.ClaimStrength,
			KillArgumentStrength: a.KillArgumentStrength,
			AuditRounds:          a.AuditRounds,
			IsDefault:            a.IsDefault,
		})
	}

	return out, nil
}
