package product

import (
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

var brandEyebrow = map[string]string{
	"openai":     "ChatGPT",
	"anthropic":  "Claude",
	"xai":        "Grok",
	"gemini":     "Gemini",
	"perplexity": "Perplexity",
	"cursor":     "Cursor",
}

func buildProductDetailContent(slug string, p *entity.Products, item response.ProductItemResp) response.ProductDetailContentResp {
	brand := brandEyebrow[p.ProductsCategoryName]
	if brand == "" {
		brand = "AI 套餐"
	}
	heroTitle := strings.TrimSuffix(item.CardTitle, "月卡") + "会员套餐"
	if !strings.HasSuffix(item.CardTitle, "月卡") {
		heroTitle = item.CardTitle + " 会员套餐"
	}
	heroLead := item.CardSubtitle + " 支付宝 / 微信自助下单，约 1 分钟开通；额度与订单可在会员中心查看。"
	bullets := appendShareFeatureBullet(append([]string(nil), item.CardFeatures...), item.ShareSeats)

	base := response.ProductDetailContentResp{
		Eyebrow:     fmt.Sprintf("%s · %s", brand, p.SKUProductName),
		HeroTitle:   heroTitle,
		HeroLead:    heroLead,
		HeroBullets: bullets,
		HeroTags:    []string{"支付宝 / 微信", "无需海外卡", "自助约 1 分钟"},
		Audiences: []response.ProductDetailAudienceResp{
			{Title: "个人用户", Desc: "希望低于官网价、国内支付自助开通。"},
			{Title: "日常办公", Desc: "写作、翻译、资料整理等高频使用。"},
			{Title: "开发者", Desc: "需要稳定额度与明确套餐边界。"},
			{Title: "小团队", Desc: "可先单账号试用，批量走企业采购。"},
		},
		Steps: []response.ProductDetailStepResp{
			{Title: "选择套餐", Desc: "确认本页套餐与价格，点击立即购买。"},
			{Title: "登录并支付", Desc: "使用支付宝或微信完成付款。"},
			{Title: "开始使用", Desc: "约 1 分钟到账，在会员中心查看额度与订单。"},
		},
		CtaTitle:    "开通 " + p.SKUProductName,
		CtaSubtitle: "选好套餐并完成支付后，在会员中心查看额度与使用说明。",
		Faqs: []response.ProductDetailFaqResp{
			{Q: "国内能开通吗？需要海外卡吗？", A: "可以。支持支付宝 / 微信，无需海外信用卡。"},
			{Q: "多久到账？", A: "通常约 1 分钟自动开通，可在会员中心查看状态。"},
			{Q: "失败怎么办？", A: "支付异常或未到账请联系在线客服，核实后按规则处理。"},
			{Q: "和官网套餐有什么关系？", A: "本站为独立第三方 AI 服务平台，提供会员套餐与开通服务，与商标持有人无隶属关系。"},
		},
		Related: []response.ProductDetailRelatedResp{
			{Label: "查看全部套餐", Slug: canonicalSlugForSKU(p.SKUCode)},
		},
	}

	applyDetailOverrides(slug, &base)
	return base
}

func appendShareFeatureBullet(bullets []string, seats int) []string {
	if seats <= 0 {
		seats = 1
	}
	line := fmt.Sprintf("支持%d人共用", seats)
	for _, b := range bullets {
		if strings.Contains(b, "人共用") {
			return bullets
		}
	}
	return append(bullets, line)
}

func applyDetailOverrides(slug string, d *response.ProductDetailContentResp) {
	switch slug {
	case "chatgpt-plus":
		d.Eyebrow = "ChatGPT · 主力档"
		d.HeroTitle = "ChatGPT Plus 会员套餐"
		d.HeroLead = "主力模型组合，适合日常对话、写作与轻量开发。支付宝 / 微信自助下单，无需海外信用卡，约 1 分钟开通后在会员中心查看额度。"
		d.CompareTitle = "Plus 还是 Pro 档位？"
		d.CompareBody = "日常个人使用 Plus 性价比最高；需要更高 RPM、更多 tokens 或 Codex 重度场景，可选 Pro 5X / 20X 档位。"
		d.Related = []response.ProductDetailRelatedResp{
			{Label: "GPT Go 入门", Slug: "gpt-go"},
			{Label: "ChatGPT Pro 5X", Slug: "chatgpt-pro-5x"},
		}
	case "cursor-pro-plus":
		d.Eyebrow = "Cursor · 重度版"
		d.HeroTitle = "Cursor Pro+ 会员套餐"
		d.HeroLead = "面向重度编程与大型重构，额度约为 Pro 的 3 倍。支持前沿模型与 Max Mode，Cloud Agents、MCP 等能力按套餐说明配置。支付宝 / 微信自助下单，约 1 分钟到账。"
		d.HeroBullets = []string{
			"约 3 倍 Pro 的 Agent 额度，减少触顶",
			"前沿模型 + Max Mode 重度可用",
			"Cloud Agents 并行 + MCP / skills / hooks",
			"支付宝 / 微信自助支付，约 1 分钟开通",
		}
		d.Audiences = []response.ProductDetailAudienceResp{
			{Title: "重度编程", Desc: "长时间使用前沿模型 / Max Mode，减少撞额度。"},
			{Title: "大型重构", Desc: "整库理解、多文件改写更从容。"},
			{Title: "常触顶的 Pro 用户", Desc: "一周多次触顶，升级后立刻缓解。"},
			{Title: "高强度个人", Desc: "强度高但尚未到全天跑 Agent 的量级。"},
		}
		d.CompareTitle = "Pro+ 还是 Ultra？"
		d.CompareBody = "若一周多次触顶、重度使用前沿模型 / Max Mode，Pro+（约 3× 额度）通常最划算。几乎全天跑 Cloud Agents、多仓库并行或团队场景，再考虑 Ultra（约 20×）。日常轻度使用可继续 Pro 更省。"
		d.Related = []response.ProductDetailRelatedResp{
			{Label: "Cursor Pro 套餐", Slug: "cursor-pro"},
			{Label: "查看全部 Cursor 套餐", Slug: "cursor-pro"},
		}
	}
}
