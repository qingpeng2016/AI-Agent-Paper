package arxiv

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/conf"
	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-agent-paper/domain/http/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
)

var idFromURLRe = regexp.MustCompile(`arxiv\.org/abs/([^/]+)`)

// Client 实现 domain/http/repository.ArxivRepo
type Client struct {
	http    *httpx.Client
	baseURL string
}

func NewClient(http *httpx.Client, sources repository.PaperLiteraturePlatformRepo) httprepo.ArxivRepo {
	baseURL := httpx.ResolveLiteratureBaseURL(
		context.Background(),
		sources,
		httpx.LiteratureSourceCodeArxiv,
		conf.DefaultLiteratureArxivBaseURL,
	)
	return &Client{http: http, baseURL: baseURL}
}

func (c *Client) Search(ctx context.Context, q httpentity.ArxivSearchQuery) ([]httpentity.ArxivPaper, error) {
	papers, _, err := c.searchFeed(ctx, q)
	return papers, err
}

func (c *Client) Count(ctx context.Context, query string) (int, error) {
	_, total, err := c.searchFeed(ctx, httpentity.ArxivSearchQuery{Query: query, MaxResults: 1, Start: 0})
	return total, err
}

func (c *Client) searchFeed(ctx context.Context, q httpentity.ArxivSearchQuery) ([]httpentity.ArxivPaper, int, error) {
	searchQuery := url.QueryEscape("all:" + q.Query)
	start := q.Start
	if start < 0 {
		start = 0
	}
	maxResults := q.MaxResults
	if maxResults <= 0 {
		maxResults = 1
	}
	apiURL := fmt.Sprintf("%s?search_query=%s&start=%d&max_results=%d", c.baseURL, searchQuery, start, maxResults)
	resp, err := c.http.Get(ctx, apiURL, nil, nil)
	if err != nil {
		return nil, 0, httpx.LogLiteratureTransportError(ctx, "arxiv", "search", apiURL, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, 0, httpx.NewLiteratureHTTPError(ctx, "arxiv", "search", apiURL, resp, fmt.Errorf("arxiv http %d", resp.StatusCode()))
	}
	var feed feedXML
	if err := xml.Unmarshal(resp.Body(), &feed); err != nil {
		return nil, 0, httpx.NewLiteratureHTTPError(ctx, "arxiv", "search", apiURL, resp, err)
	}
	total := feed.TotalResults
	out := make([]httpentity.ArxivPaper, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		p := entryToPaper(e)
		if p.Title != "" {
			out = append(out, p)
		}
	}
	return out, total, nil
}

type feedXML struct {
	TotalResults int        `xml:"http://a9.com/-/spec/opensearch/1.1/ totalResults"`
	Entries      []entryXML `xml:"entry"`
}

type entryXML struct {
	ID        string      `xml:"id"`
	Title     string      `xml:"title"`
	Summary   string      `xml:"summary"`
	Published string      `xml:"published"`
	Authors   []authorXML `xml:"author"`
}

type authorXML struct {
	Name string `xml:"name"`
}

func entryToPaper(e entryXML) httpentity.ArxivPaper {
	title := strings.TrimSpace(strings.Join(strings.Fields(e.Title), " "))
	abstract := strings.TrimSpace(strings.Join(strings.Fields(e.Summary), " "))
	ext := e.ID
	if m := idFromURLRe.FindStringSubmatch(e.ID); len(m) == 2 {
		ext = "arxiv:" + m[1]
	}
	var authors []string
	for _, a := range e.Authors {
		n := strings.TrimSpace(a.Name)
		if n != "" {
			authors = append(authors, n)
		}
	}
	year := 0
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(e.Published)); err == nil {
		year = t.Year()
	}
	return httpentity.ArxivPaper{
		ExternalKey:   ext,
		Title:         title,
		Authors:       authors,
		Abstract:      abstract,
		PublishedYear: year,
		URL:           strings.TrimSpace(e.ID),
	}
}
