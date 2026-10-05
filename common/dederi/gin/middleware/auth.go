package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qingpeng2016/ai-token-mall/common/auth"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
)

const ContextUserIDKey = "userID"

func authTokenFromRequest(c *gin.Context) string {
	if raw := strings.TrimSpace(c.GetHeader("Authorization")); raw != "" {
		return strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
	}
	if cookie, err := c.Cookie(constants.AuthTokenCookie); err == nil {
		return strings.TrimSpace(cookie)
	}
	return ""
}

// TryUserIDFromRequest 可选登录：有合法 token 则返回 userID，否则 ok=false
func TryUserIDFromRequest(c *gin.Context) (uint, bool) {
	token := authTokenFromRequest(c)
	if token == "" {
		return 0, false
	}
	uid, err := auth.ParseUserToken(token, auth.SessionTokenSecret)
	if err != nil {
		return 0, false
	}
	return uid, true
}

// RequireAuth 解析 Authorization: Bearer 或 Cookie atm_token
func RequireAuth(c *gin.Context) {
	token := authTokenFromRequest(c)
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	uid, err := auth.ParseUserToken(token, auth.SessionTokenSecret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "unauthorized", "data": nil})
		return
	}
	c.Set(ContextUserIDKey, uid)
	c.Next()
}

func UserIDFromContext(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}
