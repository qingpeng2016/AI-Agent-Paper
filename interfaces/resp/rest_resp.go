package resp

type ApiResp struct {
	Code    int64       `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

type PaginationData struct {
	Total int64 `json:"total"` //总数
	Start int32 `json:"start"` //起始值
	Limit int32 `json:"limit"` //间隔数
}
