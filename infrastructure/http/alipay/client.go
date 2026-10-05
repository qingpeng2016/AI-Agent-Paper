package alipay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/conf"
	httpentity "github.com/qingpeng2016/ai-token-mall/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-token-mall/domain/http/repository"
	httpx "github.com/qingpeng2016/ai-token-mall/infrastructure/http"
)

const methodAccountLogQuery = "alipay.data.bill.accountlog.query"

// Client 实现 domain/http/repository.AlipayRepo
type Client struct {
	http *httpx.Client
}

func NewClient(http *httpx.Client) httprepo.AlipayRepo {
	return &Client{http: http}
}

func (c *Client) QueryAccountLogs(ctx context.Context, q httpentity.AccountLogQuery) (*httpentity.AccountLogQueryResult, error) {
	cfg := conf.GetAlipayConf()
	if cfg == nil || cfg.AppID == "" || cfg.PrivateKey == "" {
		return nil, fmt.Errorf("alipay config not configured")
	}
	pageNo := q.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 2000
	}
	biz := map[string]interface{}{
		"start_time": q.StartTime,
		"end_time":   q.EndTime,
		"page_no":    fmt.Sprintf("%d", pageNo),
		"page_size":  fmt.Sprintf("%d", pageSize),
	}
	bizJSON, _ := json.Marshal(biz)

	params := map[string]string{
		"app_id":      cfg.AppID,
		"method":      methodAccountLogQuery,
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(bizJSON),
	}

	sign, err := signParams(params, cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	params["sign"] = sign

	gateway := cfg.Gateway
	if gateway == "" {
		gateway = "https://openapi.alipay.com/gateway.do"
	}
	resp, err := c.http.PostForm(ctx, gateway, params)
	if err != nil {
		return nil, err
	}
	body := resp.String()
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &wrapper); err != nil {
		return nil, fmt.Errorf("parse alipay response: %w", err)
	}
	raw, ok := wrapper["alipay_data_bill_accountlog_query_response"]
	if !ok {
		return nil, fmt.Errorf("unexpected alipay response: %s", body)
	}
	var result httpentity.AccountLogQueryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if result.Code != "10000" {
		return &result, fmt.Errorf("alipay error: %s %s (%s %s)", result.Code, result.Msg, result.SubCode, result.SubMsg)
	}
	return &result, nil
}

func signParams(params map[string]string, privateKeyPEM string) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(params[k])
	}
	priv, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(b.String()))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("invalid private key pem")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not rsa private key")
	}
	return rsaKey, nil
}
