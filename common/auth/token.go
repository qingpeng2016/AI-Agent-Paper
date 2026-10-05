package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

// SessionTokenSecret / SessionTokenTTL 与登录签发一致（MVP 固定密钥，生产应走配置）
const (
	SessionTokenSecret = "ai-token-mall-session-signing-key"
	SessionTokenTTL    = 720 * time.Hour
)

// IssueUserToken 签发登录 token（payload: userID:expUnix:hmac）
func IssueUserToken(userID uint, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("empty auth secret")
	}
	exp := time.Now().Add(ttl).Unix()
	payload := fmt.Sprintf("%d:%d", userID, exp)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	raw := payload + ":" + sig
	return base64.RawURLEncoding.EncodeToString([]byte(raw)), nil
}

// ParseUserToken 解析并校验 token，返回 userID
func ParseUserToken(token, secret string) (uint, error) {
	dec, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, ErrInvalidToken
	}
	parts := strings.Split(string(dec), ":")
	if len(parts) != 3 {
		return 0, ErrInvalidToken
	}
	uid64, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	if time.Now().Unix() > exp {
		return 0, ErrInvalidToken
	}
	payload := parts[0] + ":" + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return 0, ErrInvalidToken
	}
	return uint(uid64), nil
}
