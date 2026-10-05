package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/trace"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func TraceRequestLog(c *gin.Context) {
	// 记录请求开始时间
	start := time.Now()

	// 设置trace id
	if c.GetHeader(trace.HeaderTraceID) == "" {
		c.Set(trace.TraceID, trace.GenerateTraceId())
	} else {
		c.Set(trace.TraceID, c.GetHeader(trace.HeaderTraceID))
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// 将请求体重置为原始状态
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// 设置返回header
	c.Header(trace.HeaderTraceID, trace.GetTraceIdByCtx(c))

	// 继续执行后续的中间件和路由处理
	c.Next()

	// 打印请求数据
	logger.InfoZ(c, "request log",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.RequestURI),
		zap.Any("header", c.Request.Header),
		zap.String("request_body", string(body)),
		zap.Int64("duration_ms", time.Since(start).Milliseconds()),
	)
}

// CORSMiddleware sets the CORS headers on the response.
// Our custom CORS middleware.
func CORSMiddleware(c *gin.Context) {
	method := c.Request.Method
	header := c.Request.Header.Get("Access-Control-Request-Headers")
	if origin := c.Request.Header.Get("Origin"); origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Vary", "Origin")
	} else {
		c.Header("Access-Control-Allow-Origin", "*")
	}
	if method == "OPTIONS" {
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", header)
		c.Header("Access-Control-Max-Age", "86400")
		c.AbortWithStatus(http.StatusNoContent)
	}
	c.Next()
}
