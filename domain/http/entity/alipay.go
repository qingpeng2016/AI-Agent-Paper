// 官方参考：https://opendocs.alipay.com/open/028xq9
package entity

// AccountLogQuery 账务明细查询入参（alipay.data.bill.accountlog.query）
type AccountLogQuery struct {
	StartTime string
	EndTime   string
	PageNo    int
	PageSize  int
}

// AccountLogQueryResult OpenAPI 业务响应体
type AccountLogQueryResult struct {
	Code       string `json:"code"`
	Msg        string `json:"msg"`
	SubCode    string `json:"sub_code"`
	SubMsg     string `json:"sub_msg"`
	TotalSize  string `json:"total_size"`
	PageNo     string `json:"page_no"`
	PageSize   string `json:"page_size"`
	DetailList string `json:"detail_list"`
}
