package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

// LarkMessage Lark 消息结构
type LarkMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// LarkCardMessage Lark 卡片消息结构
type LarkCardMessage struct {
	MsgType string `json:"msg_type"`
	Card    struct {
		Elements []LarkCardElement `json:"elements"`
		Header   LarkCardHeader    `json:"header"`
	} `json:"card"`
}

// LarkCardHeader 卡片头部
type LarkCardHeader struct {
	Title LarkCardTitle `json:"title"`
}

// LarkCardTitle 卡片标题
type LarkCardTitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// LarkCardElement 卡片元素
type LarkCardElement struct {
	Tag    string          `json:"tag"`
	Text   *LarkCardText   `json:"text,omitempty"`
	Fields []LarkCardField `json:"fields,omitempty"`
}

// LarkCardText 卡片文本
type LarkCardText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// LarkCardField 卡片字段
type LarkCardField struct {
	IsShort bool         `json:"is_short"`
	Text    LarkCardText `json:"text"`
}

// LarkClient Lark 客户端
type LarkClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewLarkClient 创建 Lark 客户端
func NewLarkClient(webhookURL string) *LarkClient {
	return &LarkClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendTextMessage 发送文本消息
func (c *LarkClient) SendTextMessage(ctx context.Context, text string) error {
	message := LarkMessage{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: text,
		},
	}

	return c.sendMessage(ctx, message)
}

// SendCardMessage 发送卡片消息
func (c *LarkClient) SendCardMessage(ctx context.Context, title, content string, fields map[string]string) error {
	card := LarkCardMessage{
		MsgType: "interactive",
		Card: struct {
			Elements []LarkCardElement `json:"elements"`
			Header   LarkCardHeader    `json:"header"`
		}{
			Header: LarkCardHeader{
				Title: LarkCardTitle{
					Tag:     "plain_text",
					Content: title,
				},
			},
		},
	}

	// 添加内容文本
	card.Card.Elements = append(card.Card.Elements, LarkCardElement{
		Tag: "div",
		Text: &LarkCardText{
			Tag:     "lark_md",
			Content: content,
		},
	})

	// 添加字段（每个字段独占一行）
	if len(fields) > 0 {
		for key, value := range fields {
			card.Card.Elements = append(card.Card.Elements, LarkCardElement{
				Tag: "div",
				Text: &LarkCardText{
					Tag:     "lark_md",
					Content: fmt.Sprintf("**%s:** %s", key, value),
				},
			})
		}
	}

	return c.sendMessage(ctx, card)
}

// SendCardMessageWithOrder 发送卡片消息（保持字段顺序）
func (c *LarkClient) SendCardMessageWithOrder(ctx context.Context, title, content string, fields []FieldPair) error {
	card := LarkCardMessage{
		MsgType: "interactive",
		Card: struct {
			Elements []LarkCardElement `json:"elements"`
			Header   LarkCardHeader    `json:"header"`
		}{
			Header: LarkCardHeader{
				Title: LarkCardTitle{
					Tag:     "plain_text",
					Content: title,
				},
			},
		},
	}

	// 添加内容文本
	card.Card.Elements = append(card.Card.Elements, LarkCardElement{
		Tag: "div",
		Text: &LarkCardText{
			Tag:     "lark_md",
			Content: content,
		},
	})

	// 添加字段（每个字段独占一行，保持顺序）
	for _, field := range fields {
		card.Card.Elements = append(card.Card.Elements, LarkCardElement{
			Tag: "div",
			Text: &LarkCardText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("**%s:** %s", field.Key, field.Value),
			},
		})
	}

	return c.sendMessage(ctx, card)
}

// SendAlert 发送报警消息
func (c *LarkClient) SendAlert(ctx context.Context, alertType, service, message string, fields map[string]string) error {
	title := fmt.Sprintf("🚨 %s 报警", alertType)
	content := fmt.Sprintf("**服务:** %s\n**消息:** %s\n**时间:** %s",
		service, message, time.Now().Format("2006-01-02 15:04:05"))

	return c.SendCardMessage(ctx, title, content, fields)
}

// SendAlertWithOrder 发送报警消息（保持字段顺序）
func (c *LarkClient) SendAlertWithOrder(ctx context.Context, alertType, service, message string, fields []FieldPair) error {
	title := fmt.Sprintf("🚨 %s 报警", alertType)
	content := fmt.Sprintf("**服务:** %s\n**消息:** %s\n**时间:** %s",
		service, message, time.Now().Format("2006-01-02 15:04:05"))

	return c.SendCardMessageWithOrder(ctx, title, content, fields)
}

// SendErrorAlertWithOrder 发送错误报警（保持字段顺序）
func (c *LarkClient) SendErrorAlertWithOrder(ctx context.Context, service, errorMsg string, fields []FieldPair) error {
	return c.SendAlertWithOrder(ctx, "错误", service, errorMsg, fields)
}

// SendWarningAlertWithOrder 发送警告报警（保持字段顺序）
func (c *LarkClient) SendWarningAlertWithOrder(ctx context.Context, service, warningMsg string, fields []FieldPair) error {
	return c.SendAlertWithOrder(ctx, "警告", service, warningMsg, fields)
}

// SendInfoAlert 发送消息通知
func (c *LarkClient) SendInfoAlert(ctx context.Context, service, infoMsg string, fields []FieldPair) error {
	title := fmt.Sprintf("✅ %s 消息", "信息")
	content := fmt.Sprintf("**服务:** %s\n**消息:** %s\n**时间:** %s",
		service, infoMsg, time.Now().Format("2006-01-02 15:04:05"))

	return c.SendCardMessageWithOrder(ctx, title, content, fields)
}

// sendMessage 发送消息到 Lark
func (c *LarkClient) sendMessage(ctx context.Context, message interface{}) error {
	logger.InfoZ(ctx, "lark-sendMessage-Start",
		zap.String("webhookURL", c.webhookURL),
		zap.Any("message", message))

	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.ErrorZ(ctx, "lark-sendMessage-FailedToMarshal", zap.Error(err))
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	logger.InfoZ(ctx, "lark-sendMessage-JSONData",
		zap.String("jsonData", string(jsonData)))

	req, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.ErrorZ(ctx, "lark-sendMessage-FailedToCreateRequest", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	logger.InfoZ(ctx, "lark-sendMessage-SendingRequest",
		zap.String("url", c.webhookURL))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.ErrorZ(ctx, "lark-sendMessage-FailedToSendRequest", zap.Error(err))
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	logger.InfoZ(ctx, "lark-sendMessage-ResponseReceived",
		zap.Int("statusCode", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		// 读取响应体以获取更多错误信息
		body, _ := io.ReadAll(resp.Body)
		logger.ErrorZ(ctx, "lark-sendMessage-UnexpectedStatusCode",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("responseBody", string(body)))
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	logger.InfoZ(ctx, "lark-sendMessage-Success")
	return nil
}
