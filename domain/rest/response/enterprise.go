package response

type EnterpriseProductListResp struct {
	Products []EnterpriseProductItemResp `json:"products"`
}

type EnterpriseProductItemResp struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Badge       *string  `json:"badge,omitempty"`
	PriceHint   string   `json:"price_hint"`
	Seats       string   `json:"seats"`
	Features    []string `json:"features"`
	Tagline     string   `json:"tagline"`
	ButtonLabel string   `json:"button_label"`
	IsFeatured  bool     `json:"is_featured"`
	Sort        int      `json:"sort"`
}
