package topicdiscovery

import (
	"context"
	"errors"
	"math/rand/v2"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	"go.uber.org/zap"
)

const (
	sourceArxiv           = "arxiv"
	sourceOpenAlex        = "openalex"
	sourceSemanticScholar = "semantic_scholar"
	sourceCrossref        = "crossref"
	sourcePubMed          = "pubmed"
)

// 选题发现 · 随机文献检索：固定 5 次第三方请求，均分目标篇数。
const literatureRetrieveRandomRounds = 5

// 随机检索统一按每页 10 条推算页码，再换算 offset / OpenAlex 页码。
const literatureRandomPageSize = 10

// arXiv 等对超大 start/offset 易慢/429/500；OpenAlex 深页也易失败。
const (
	literatureRandomMaxOffset       = 9800
	literatureRandomMaxOpenAlexPage = 400
	// 随机检索只需抽样；arXiv 深 offset（如 start=8000+）极易触发限流。
	literatureRandomArxivMaxPages = 50
)

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

// literatureRandomRoundLimits：目标篇数较少时每轮最多 1 篇、多轮换源；否则均分 5 轮。
func literatureRandomRoundLimits(maxPapers int) [literatureRetrieveRandomRounds]int {
	if maxPapers <= 0 {
		var limits [literatureRetrieveRandomRounds]int
		limits[0] = 1
		return limits
	}
	if maxPapers <= literatureRetrieveRandomRounds {
		var limits [literatureRetrieveRandomRounds]int
		for i := 0; i < maxPapers; i++ {
			limits[i] = 1
		}
		return limits
	}
	return literatureRoundLimits(maxPapers)
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

// literatureRandomPageCount 由 Count 得到的 total 推算总页数（每页 literatureRandomPageSize 条）。
func literatureRandomPageCount(total int, kind literaturePagingKind, sourceCode string) int {
	if total <= 0 {
		return 1
	}
	n := (total + literatureRandomPageSize - 1) / literatureRandomPageSize
	if n < 1 {
		return 1
	}
	if strings.TrimSpace(sourceCode) == sourceArxiv && n > literatureRandomArxivMaxPages {
		n = literatureRandomArxivMaxPages
	}
	if kind == literaturePagingOpenAlexPage && n > literatureRandomMaxOpenAlexPage {
		return literatureRandomMaxOpenAlexPage
	}
	maxPageByOffset := literatureRandomMaxOffset/literatureRandomPageSize + 1
	if kind != literaturePagingOpenAlexPage && n > maxPageByOffset {
		return maxPageByOffset
	}
	return n
}

// pickUnusedRandomPageIndex 在 [minPage,maxPage] 内随机一页（offset 平台 0-based，OpenAlex 1-based）。
func pickUnusedRandomPageIndex(minPage, maxPage int, used map[int]struct{}) int {
	if used == nil {
		used = map[int]struct{}{}
	}
	if maxPage < minPage {
		maxPage = minPage
	}
	tryPick := func() (int, bool) {
		var candidate int
		if maxPage <= minPage {
			candidate = minPage
		} else {
			candidate = minPage + rand.IntN(maxPage-minPage+1)
		}
		if _, ok := used[candidate]; ok {
			return 0, false
		}
		used[candidate] = struct{}{}
		return candidate, true
	}
	for i := 0; i < 48; i++ {
		if page, ok := tryPick(); ok {
			return page
		}
	}
	for page := minPage; page <= maxPage; page++ {
		if _, ok := used[page]; ok {
			continue
		}
		used[page] = struct{}{}
		return page
	}
	page := minPage
	if maxPage > minPage {
		page = minPage + rand.IntN(maxPage-minPage+1)
	}
	used[page] = struct{}{}
	return page
}

// pickUnusedRandomSlot 先按 total 与每页 10 条得到页码范围，再随机一页并换算 API slot（offset 或 OpenAlex 页码）。
func pickUnusedRandomSlot(total int, kind literaturePagingKind, sourceCode string, used map[int]struct{}) (slot int, pageIndex int, pageCount int) {
	pageCount = literatureRandomPageCount(total, kind, sourceCode)
	switch kind {
	case literaturePagingOpenAlexPage:
		pageIndex = pickUnusedRandomPageIndex(1, pageCount, used)
		return pageIndex, pageIndex, pageCount
	default:
		maxPageIdx := pageCount - 1
		if maxPageIdx < 0 {
			maxPageIdx = 0
		}
		pageIndex = pickUnusedRandomPageIndex(0, maxPageIdx, used)
		return pageIndex * literatureRandomPageSize, pageIndex, pageCount
	}
}

func (s *TopicDiscoveryRunService) retrieveLiteratureRandom(
	ctx context.Context,
	query string,
	input topicRunInput,
	extraMeta map[string]any,
	scope literature.LiteratureRetrieveLogScope,
) ([]storedLiteratureHit, map[string]any) {
	enabled := filterRetrieveSourceCodes(input.SourceCodes)
	seen := map[string]struct{}{}
	hits := make([]storedLiteratureHit, 0, input.MaxPapers)
	if len(enabled) == 0 {
		return hits, extraMeta
	}

	roundLimits := literatureRandomRoundLimits(input.MaxPapers)
	usedSlots := map[string]map[int]struct{}{}
	countBySource := map[string]struct {
		total int
		err   error
	}{}

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

		var total int
		var countErr error
		logger.InfoZ(ctx, "topic-retrieve-random-round",
			zap.Int("round", round+1),
			zap.String("source_code", code),
			zap.Int("hits_so_far", len(hits)))

		if cached, ok := countBySource[code]; ok {
			total, countErr = cached.total, cached.err
		} else {
			total, countErr = s.litSearch.CountPlatformTotal(ctx, code, query)
			countBySource[code] = struct {
				total int
				err   error
			}{total: total, err: countErr}
		}
		if countErr != nil {
			literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, literature.LiteraturePlatformRequestLog{
				SourceCode: code,
				Op:         "count",
				Input: map[string]any{
					"query":       query,
					"round":       round + 1,
					"random_mode": true,
				},
			}, countErr)
		}
		kind := literaturePagingKindFor(code)
		slot, pageIndex, pageCount := pickUnusedRandomSlot(total, kind, code, usedSlots[code])

		reqInput := map[string]any{
			"query":       query,
			"limit":       limit,
			"round":       round + 1,
			"total":       total,
			"page_size":   literatureRandomPageSize,
			"page_count":  pageCount,
			"page_index":  pageIndex,
			"slot":        slot,
			"random_mode": true,
		}
		if countErr != nil {
			reqInput["count_error"] = countErr.Error()
		}

		searchResult, searchErr := s.litSearch.SearchPlatformPage(ctx, code, query, limit, slot)
		logEntry := literature.LiteraturePlatformRequestLog{
			SourceCode: code,
			Op:         "search",
			Input:      reqInput,
		}
		if searchErr != nil {
			literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, logEntry, searchErr)
			continue
		}
		if searchResult == nil {
			logEntry.Error = "empty search result"
			literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, logEntry, errors.New("empty search result"))
			continue
		}
		logEntry.Response = searchResult
		literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, logEntry, nil)

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
	scope literature.LiteratureRetrieveLogScope,
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
			literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, literature.LiteraturePlatformRequestLog{
				SourceCode: code,
				Op:         "search",
				Input:      map[string]any{"query": query},
				Error:      "unsupported source for retrieve search",
			}, errors.New("unsupported source for retrieve search"))
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
			logEntry := literature.LiteraturePlatformRequestLog{
				SourceCode: code,
				Op:         "search",
				Input:      reqInput,
			}
			if searchErr != nil {
				literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, logEntry, searchErr)
				break
			}
			if searchResult == nil {
				logEntry.Error = "empty search result"
				literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, logEntry, errors.New("empty search result"))
				break
			}
			logEntry.Response = searchResult
			literature.RecordLiteraturePlatformRequest(ctx, extraMeta, scope, logEntry, nil)

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
		linkURL := literature.ResolveLiteratureHitURL(code, key, h.URL, h.DOI)
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
				"pmcid":          h.PMCID,
			},
		})
	}
	return hits
}
