package tool

import (
	"crypto/rand"
	"encoding/hex"
	mathrand "math/rand"
	"strconv"
	"strings"
	"time"
)

// Random 生成随机字符串方法
// 注意: 此方法使用 math/rand，不适用于安全敏感场景（如token、密钥生成）
// 对于安全敏感场景，请使用 SecureRandom() 或 SecureRandomHex()
func Random() string {
	// 使用当前时间作为种子
	source := mathrand.NewSource(time.Now().UnixNano())
	r := mathrand.New(source)

	// 生成一个随机整数
	intervalMin := 1000000 // 最小的唯一数字范围
	intervalMax := 9999999 // 最大的唯一数字范围
	randomInt := r.Intn(mathrand.Intn(intervalMax-intervalMin+1) + intervalMin)
	ms := time.Now().Format("060102150405.000") // 北京时间
	return strings.Replace(ms, ".", "", -1) + "" + strconv.Itoa(randomInt)
}

// SecureRandomHex 生成加密安全的随机十六进制字符串
// 参数 n: 生成的字节数（返回的十六进制字符串长度为 n*2）
// 适用于: token、API key、密钥等安全敏感场景
func SecureRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SecureRandom 生成加密安全的随机字符串（带时间戳）
// 返回: 时间戳 + 16字节随机十六进制字符串
// 适用于: 需要唯一性和安全性的ID生成
func SecureRandom() (string, error) {
	randomHex, err := SecureRandomHex(16)
	if err != nil {
		return "", err
	}
	ms := time.Now().Format("060102150405.000")
	return strings.Replace(ms, ".", "", -1) + randomHex, nil
}

// SecureRandomBytes 生成加密安全的随机字节
// 参数 n: 生成的字节数
func SecureRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	return bytes, err
}
