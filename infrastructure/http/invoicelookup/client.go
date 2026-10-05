package invoicelookup

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/conf"
)

type LookupItem struct {
	Title       string
	TaxNo       string
	BankName    string
	BankAccount string
	Address     string
	Phone       string
}

type Client struct {
	httpClient     *http.Client
	chengshuAPIKey string
	mgtvBaseURL    string
}

func NewClient(cfg *conf.Config) *Client {
	timeout := conf.GetHTTPTimeout()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	skipTLS := false
	chengshuKey := strings.TrimSpace(os.Getenv("CHENGSHU_API_KEY"))
	mgtvURL := "https://tools.mgtv100.com/external/v1/invoiceInfo"
	if cfg != nil && cfg.InvoiceLookupConf != nil {
		if k := strings.TrimSpace(cfg.InvoiceLookupConf.ChengshuAPIKey); k != "" {
			chengshuKey = k
		}
		if u := strings.TrimSpace(cfg.InvoiceLookupConf.MgtvBaseURL); u != "" {
			mgtvURL = u
		}
		skipTLS = cfg.InvoiceLookupConf.InsecureSkipTLSVerify
	}
	if skipTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // third-party cert issues in dev
	}
	return &Client{
		httpClient:     &http.Client{Timeout: timeout, Transport: transport},
		chengshuAPIKey: chengshuKey,
		mgtvBaseURL:    mgtvURL,
	}
}

// Lookup 按企业名称或统一社会信用代码/税号查询开票抬头信息。
func (c *Client) Lookup(ctx context.Context, keyword string) ([]LookupItem, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("empty keyword")
	}

	var out []LookupItem
	var lastErr error

	if c.chengshuAPIKey != "" {
		items, err := c.lookupChengshu(ctx, keyword)
		if err == nil && len(items) > 0 {
			return items, nil
		}
		if err != nil {
			lastErr = err
		}
	}

	items, err := c.lookupMgtv(ctx, keyword)
	if err == nil && len(items) > 0 {
		return items, nil
	}
	if err != nil {
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return out, nil
}

func (c *Client) lookupChengshu(ctx context.Context, keyword string) ([]LookupItem, error) {
	u, err := url.Parse("https://market.gzchengshu.com/api/enterprise/shcx")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("key", c.chengshuAPIKey)
	q.Set("keyword", keyword)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var parsed chengshuResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Data.Result != "1" || parsed.Data.Detail.CompanyName == "" {
		return nil, fmt.Errorf("chengshu: no data")
	}
	d := parsed.Data.Detail
	tax := firstNonEmpty(d.TaxNumber, d.CreditCode)
	return []LookupItem{{
		Title:   d.CompanyName,
		TaxNo:   tax,
		Address: d.Address,
		Phone:   d.Phone,
	}}, nil
}

func (c *Client) lookupMgtv(ctx context.Context, keyword string) ([]LookupItem, error) {
	base := strings.TrimRight(c.mgtvBaseURL, "/")
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("companyName", keyword)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if len(body) > 0 && body[0] == '{' {
		var errWrap struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errWrap) == nil && errWrap.Error != "" {
			return nil, fmt.Errorf("mgtv: %s", errWrap.Error)
		}
	}

	var parsed mgtvResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	lst := parsed.Data.AutoCompleteInfoLst
	if len(lst) == 0 {
		return nil, fmt.Errorf("mgtv: no data")
	}
	items := make([]LookupItem, 0, len(lst))
	for _, row := range lst {
		items = append(items, LookupItem{
			Title:       row.CorpName,
			TaxNo:       row.TaxpayerNum,
			BankName:    row.BankName,
			BankAccount: row.BankAccount,
			Address:     row.Address,
			Phone:       row.Telephone,
		})
	}
	return items, nil
}

type chengshuResp struct {
	Data struct {
		Result string `json:"result"`
		Detail struct {
			CompanyName string `json:"companyName"`
			CreditCode  string `json:"creditCode"`
			TaxNumber   string `json:"taxNumber"`
			Address     string `json:"address"`
			Phone       string `json:"phone"`
		} `json:"detail"`
	} `json:"data"`
}

type mgtvResp struct {
	Data struct {
		AutoCompleteInfoLst []struct {
			CorpName     string `json:"corpName"`
			TaxpayerNum  string `json:"taxpayerNum"`
			Address      string `json:"address"`
			Telephone    string `json:"telephone"`
			BankName     string `json:"bankName"`
			BankAccount  string `json:"bankAccount"`
		} `json:"autoCompleteInfoLst"`
	} `json:"data"`
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
