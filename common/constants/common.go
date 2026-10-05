package constants

import "time"

var (
	// AppLocation 业务时区（北京时间 Asia/Shanghai, UTC+8）
	AppLocation, _ = time.LoadLocation("Asia/Shanghai")
)
