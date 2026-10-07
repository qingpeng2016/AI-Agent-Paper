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

	"github.com/qingpeng2016/ai-token-mall/conf"
	httpentity "github.com/qingpeng2016/ai-token-mall/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-token-mall/domain/http/repository"
	httpx "github.com/qingpeng2016/ai-token-mall/infrastructure/http"
)

var idFromURLRe = regexp.MustCompile(`arxiv\.org/abs/([^/]+)`)

// Client 实现 domain/http/repository.ArxivRepo
type Client struct {
	http    *httpx.Client
	baseURL string
}

func NewClient(http *httpx.Client, cfg *conf.Config) httprepo.ArxivRepo {
	lit := conf.GetLiteratureConf()
	baseURL := conf.DefaultLiteratureArxivBaseURL
	if lit != nil && lit.Arxiv != nil && strings.TrimSpace(lit.Arxiv.BaseURL) != "" {
		baseURL = strings.TrimSpace(lit.Arxiv.BaseURL)
	}
	return &Client{http: http, baseURL: baseURL}
}

func (c *Client) Search(ctx context.Context, q httpentity.ArxivSearchQuery) ([]httpentity.ArxivPaper, error) {
	searchQuery := url.QueryEscape("all:" + q.Query)
	apiURL := fmt.Sprintf("%s?search_query=%s&start=0&max_results=%d", c.baseURL, searchQuery, q.MaxResults)
	resp, err := c.http.Get(ctx, apiURL, nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("arxiv http %d", resp.StatusCode())
	}
	var feed feedXML
	if err := xml.Unmarshal(resp.Body(), &feed); err != nil {
		return nil, err
	}
	out := make([]httpentity.ArxivPaper, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		p := entryToPaper(e)
		if p.Title != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

type feedXML struct {
	Entries []entryXML `xml:"entry"`
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
