package response

type NavMenuResp struct {
	MegaMenu    []NavMegaColumnResp   `json:"mega_menu"`
	BrandMenus  []NavBrandMenuResp    `json:"brand_menus"`
}

type NavMegaColumnResp struct {
	Title string           `json:"title"`
	Items []NavPlanItemResp  `json:"items"`
}

type NavBrandMenuResp struct {
	ID     string            `json:"id"`
	Label  string            `json:"label"`
	HotTagName string          `json:"hot_tag_name,omitempty"`
	Items  []NavPlanItemResp `json:"items"`
}

type NavPlanItemResp struct {
	Label    string `json:"label"`
	Price    string `json:"price"`
	Href     string `json:"href"`
	Featured bool   `json:"featured,omitempty"`
}
