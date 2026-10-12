package paper

import (
	"context"
	"math/rand/v2"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

// 选题发现 · 随机文献检索：固定 5 次第三方请求，均分目标篇数。
const literatureRetrieveRandomRounds = 5

// 顺序检索：每页条数；返回少于此数即视为该平台在该关键词下已无更多结果。
const literatureSequentialPageSize = 10

var literatureRetrieveSupportedSources = map[string]struct{}{
	sourceArxiv:           {},
	sourceOpenAlex:        {},
	sourceSemanticScholar: {},
	sourceCrossref:        {},
	sourcePubMed:          {},
}

// literatureRoundLimits 将 maxPapers 均分 5 轮；不能整除的余数全部加在第 1 轮。
func literatureRoundLimits(maxPapers int) [literatureRetrieveRandomRounds]int {
	var limits [literatureRetrieveRandomRounds]int
	if maxPapers <= 0 {
		limits[0] = 1
		return limits
	}
	base := maxPapers / literatureRetrieveRandomRounds
	rem := maxPapers % literatureRetrieveRandomRounds
	for i := range limits {
		limits[i] = base
	}
	limits[0] += rem
	return limits
}

func filterRetrieveSourceCodes(codes []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, c := range codes {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, ok := literatureRetrieveSupportedSources[c]; !ok {
			continue
		}
		if _, dup := seen[c]; dup {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}

type literaturePagingKind int

const (
	literaturePagingOffset literaturePagingKind = iota
	literaturePagingOpenAlexPage
)

func literaturePagingKindFor(code string) literaturePagingKind {
	if code == sourceOpenAlex {
		return literaturePagingOpenAlexPage
	}
	return literaturePagingOffset
}

// pickUnusedRandomSlot 在已知总数后随机分页位置；同一平台已用过的 slot 不再使用。
func pickUnusedRandomSlot(total, pageSize int, kind literaturePagingKind, used map[int]struct{}) int {
	if used == nil {
		used = map[int]struct{}{}
	}
	if pageSize <= 0 {
		pageSize = 1
	}

	var minSlot, maxSlot int
	switch kind {
	case literaturePagingOpenAlexPage:
		minSlot = 1
		if total <= 0 {
			maxSlot = 1
		} else {
			maxSlot = (total + pageSize - 1) / pageSize
			if maxSlot < 1 {
				maxSlot = 1
			}
		}
	default:
		minSlot = 0
		if total <= 0 {
			maxSlot = 0
		} else {
			maxSlot = total - pageSize
			if maxSlot < 0 {
				maxSlot = 0
			}
		}
	}

	tryPick := func() (int, bool) {
		var candidate int
		if maxSlot <= minSlot {
			candidate = minSlot
		} else if kind == literaturePagingOpenAlexPage {
			candidate = minSlot + rand.IntN(maxSlot-minSlot+1)
		} else {
			candidate = minSlot + rand.IntN(maxSlot-minSlot+1)
		}
		if _, ok := used[candidate]; ok {
			return 0, false
		}
		used[candidate] = struct{}{}
		return candidate, true
	}

	for i := 0; i < 48; i++ {
		if slot, ok := tryPick(); ok {
			return slot
		}
	}
	for slot := minSlot; slot <= maxSlot; slot++ {
		if _, ok := used[slot]; ok {
			continue
		}
		used[slot] = struct{}{}
		return slot
	}
	slot := minSlot
	if kind == literaturePagingOpenAlexPage && maxSlot > minSlot {
		slot = minSlot + rand.IntN(maxSlot-minSlot+1)
	} else if maxSlot > minSlot {
		slot = minSlot + rand.IntN(maxSlot-minSlot+1)
	}
	used[slot] = struct{}{}
	return slot
}

func (s *TopicDiscoveryRunService) retrieveLiteratureRandom(
	ctx context.Context,
	query string,
	input topicRunInput,
	extraMeta map[string]any,
) ([]storedLiteratureHit, map[string]any) {
	enabled := filterRetrieveSourceCodes(input.SourceCodes)
	seen := map[string]struct{}{}
	hits := make([]storedLiteratureHit, 0, input.MaxPapers)
	if len(enabled) == 0 {
		return hits, extraMeta
	}

	roundLimits := literatureRoundLimits(input.MaxPapers)
	usedSlots := map[string]map[int]struct{}{}

	for round := 0; round < literatureRetrieveRandomRounds && len(hits) < input.MaxPapers; round++ {
		code := enabled[rand.IntN(len(enabled))]
		limit := roundLimits[round]
		need := input.MaxPapers - len(hits)
		if limit > need {
			limit = need
		}
		if limit <= 0 {
			continue
		}

		if usedSlots[code] == nil {
			usedSlots[code] = map[int]struct{}{}
		}

		total, countErr := s.litSearch.CountPlatformTotal(ctx, code, query)
		slot := pickUnusedRandomSlot(total, limit, literaturePagingKindFor(code), usedSlots[code])

		reqInput := map[string]any{
			"query":       query,
			"limit":       limit,
			"round":       round + 1,
			"total":       total,
			"slot":        slot,
			"random_mode": true,
		}
		if countErr != nil {
			reqInput["count_error"] = countErr.Error()
		}

		searchResult, searchErr := s.litSearch.SearchPlatformPage(ctx, code, query, limit, slot)
		logEntry := LiteraturePlatformRequestLog{
			SourceCode: code,
			Input:      reqInput,
		}
		if searchErr != nil {
			logEntry.Error = searchErr.Error()
			AppendLiteraturePlatformRequest(extraMeta, logEntry)
			continue
		}
		if searchResult == nil {
			logEntry.Error = "empty search result"
			AppendLiteraturePlatformRequest(extraMeta, logEntry)
			continue
		}
		logEntry.Response = searchResult
		AppendLiteraturePlatformRequest(extraMeta, logEntry)

		srcRow, _ := s.litSource.FindActiveByCode(ctx, code)
		var sourceID uint
		if srcRow != nil {
			sourceID = srcRow.ID
		}
		hits = appendLiteratureSearchHits(hits, seen, searchResult, code, sourceID, query, input.MaxPapers)
	}
	return hits, extraMeta
}

func literatureSequentialInitialSlot(sourceCode string) int {
	if sourceCode == sourceOpenAlex {
		return 1
	}
	return 0
}

func literatureSequentialAdvanceSlot(sourceCode string, slot int) int {
	if sourceCode == sourceOpenAlex {
		return slot + 1
	}
	return slot + literatureSequentialPageSize
}

func (s *TopicDiscoveryRunService) retrieveLiteratureSequential(
	ctx context.Context,
	query string,
	input topicRunInput,
	extraMeta map[string]any,
) ([]storedLiteratureHit, map[string]any) {
	seen := map[string]struct{}{}
	hits := make([]storedLiteratureHit, 0, input.MaxPapers)
	for _, code := range s.orderSourceCodesBySort(ctx, input.SourceCodes) {
		if len(hits) >= input.MaxPapers {
			break
		}
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := literatureRetrieveSupportedSources[code]; !ok {
			AppendLiteraturePlatformRequest(extraMeta, LiteraturePlatformRequestLog{
				SourceCode: code,
				Input:      map[string]any{"query": query},
				Error:      "unsupported source for retrieve search",
			})
			continue
		}
		srcRow, _ := s.litSource.FindActiveByCode(ctx, code)
		var sourceID uint
		if srcRow != nil {
			sourceID = srcRow.ID
		}

		slot := literatureSequentialInitialSlot(code)
		for page := 1; len(hits) < input.MaxPapers; page++ {
			limit := literatureSequentialPageSize
			reqInput := map[string]any{
				"query":            query,
				"limit":            limit,
				"slot":             slot,
				"page":             page,
				"sequential_mode":  true,
			}
			searchResult, searchErr := s.litSearch.SearchPlatformPage(ctx, code, query, limit, slot)
			logEntry := LiteraturePlatformRequestLog{
				SourceCode: code,
				Input:      reqInput,
			}
			if searchErr != nil {
				logEntry.Error = searchErr.Error()
				AppendLiteraturePlatformRequest(extraMeta, logEntry)
				break
			}
			if searchResult == nil {
				logEntry.Error = "empty search result"
				AppendLiteraturePlatformRequest(extraMeta, logEntry)
				break
			}
			logEntry.Response = searchResult
			AppendLiteraturePlatformRequest(extraMeta, logEntry)

			before := len(hits)
			hits = appendLiteratureSearchHits(hits, seen, searchResult, code, sourceID, query, input.MaxPapers)
			returned := len(searchResult.Hits)
			if returned < literatureSequentialPageSize {
				break
			}
			if len(hits) == before {
				// 本页有结果但均已重复，避免死循环
				break
			}
			slot = literatureSequentialAdvanceSlot(code, slot)
		}
	}
	return hits, extraMeta
}

func appendLiteratureSearchHits(
	hits []storedLiteratureHit,
	seen map[string]struct{},
	searchResult *response.PaperLiteratureSearchResult,
	code string,
	sourceID uint,
	query string,
	maxPapers int,
) []storedLiteratureHit {
	for _, h := range searchResult.Hits {
		if len(hits) >= maxPapers {
			break
		}
		key := h.ExternalKey
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		linkURL := resolveLiteratureHitURL(code, key, h.URL, h.DOI)
		hits = append(hits, storedLiteratureHit{
			SourceID:      sourceID,
			SourceCode:    code,
			ExternalKey:   key,
			Title:         h.Title,
			Authors:       h.Authors,
			URL:           linkURL,
			DOI:           h.DOI,
			PublishedYear: h.PublishedYear,
			QueryText:     query,
			Meta: map[string]any{
				"abstract":       h.Abstract,
				"doi":            h.DOI,
				"url":            linkURL,
				"published_year": h.PublishedYear,
			},
		})
	}
	return hits
}
