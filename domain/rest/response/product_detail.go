package response

type ProductDetailResp struct {
	Slug    string                  `json:"slug"`
	Product ProductItemResp         `json:"product"`
	Detail  ProductDetailContentResp `json:"detail"`
}

type ProductDetailContentResp struct {
	Eyebrow      string                      `json:"eyebrow"`
	HeroTitle    string                      `json:"hero_title"`
	HeroLead     string                      `json:"hero_lead"`
	HeroBullets  []string                    `json:"hero_bullets"`
	HeroTags     []string                    `json:"hero_tags"`
	Audiences    []ProductDetailAudienceResp `json:"audiences"`
	CompareTitle string                      `json:"compare_title,omitempty"`
	CompareBody  string                      `json:"compare_body,omitempty"`
	Steps        []ProductDetailStepResp     `json:"steps"`
	CtaTitle     string                      `json:"cta_title"`
	CtaSubtitle  string                      `json:"cta_subtitle"`
	Faqs         []ProductDetailFaqResp      `json:"faqs"`
	Related      []ProductDetailRelatedResp  `json:"related"`
}

type ProductDetailAudienceResp struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type ProductDetailStepResp struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type ProductDetailFaqResp struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type ProductDetailRelatedResp struct {
	Label string `json:"label"`
	Slug  string `json:"slug"`
}
