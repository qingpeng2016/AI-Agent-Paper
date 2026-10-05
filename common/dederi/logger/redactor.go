package logger

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"

	"go.uber.org/zap/zapcore"
)

// SensitiveFieldNames 敏感字段名称列表（不区分大小写）
var SensitiveFieldNames = []string{
	"password",
	"passwd",
	"pwd",
	"secret",
	"api_key",
	"apikey",
	"api_secret",
	"apisecret",
	"token",
	"auth_token",
	"authtoken",
	"private_key",
	"privatekey",
	"authorization",
	"auth",
	"credential",
	"access_key",
	"accesskey",
	"access_token",
	"accesstoken",
}

// SensitiveFieldPatterns 敏感字段的正则表达式模式（更精确的匹配）
var SensitiveFieldPatterns = []*regexp.Regexp{
	// 密码相关：password, passwd, pwd, user_password 等
	regexp.MustCompile(`(?i)^.*(password|passwd|pwd).*$`),
	// 密钥相关：api_key, private_key, access_key 等（但不匹配 normal_key, public_key）
	regexp.MustCompile(`(?i)^.*(api[_-]?key|private[_-]?key|secret[_-]?key|access[_-]?key|auth[_-]?key).*$`),
	// 令牌相关：token, auth_token, access_token 等
	regexp.MustCompile(`(?i)^.*(token|auth)$`),
	regexp.MustCompile(`(?i)^.*(auth[_-]?token|access[_-]?token|bearer[_-]?token).*$`),
	// 秘密相关：secret, api_secret 等
	regexp.MustCompile(`(?i)^.*(secret).*$`),
	// 凭证相关
	regexp.MustCompile(`(?i)^.*(credential).*$`),
	// 授权相关
	regexp.MustCompile(`(?i)^(authorization)$`),
}

const RedactedValue = "***REDACTED***"

// isSensitiveField 判断字段名是否为敏感字段
func isSensitiveField(fieldName string) bool {
	lowerFieldName := strings.ToLower(fieldName)
	
	// 先检查精确匹配（优先级最高）
	for _, sensitive := range SensitiveFieldNames {
		if lowerFieldName == strings.ToLower(sensitive) {
			return true
		}
	}
	
	// 检查模式匹配
	for _, pattern := range SensitiveFieldPatterns {
		if pattern.MatchString(lowerFieldName) {
			return true
		}
	}
	
	return false
}

// redactValue 脱敏处理值
func redactValue(key string, value interface{}) interface{} {
	if !isSensitiveField(key) {
		return value
	}
	return RedactedValue
}

// redactMap 递归脱敏 map
func redactMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		if isSensitiveField(k) {
			result[k] = RedactedValue
			continue
		}
		
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = redactMap(val)
		case []interface{}:
			result[k] = redactSlice(val)
		default:
			result[k] = v
		}
	}
	return result
}

// redactSlice 递归脱敏 slice
func redactSlice(s []interface{}) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		switch val := v.(type) {
		case map[string]interface{}:
			result[i] = redactMap(val)
		case []interface{}:
			result[i] = redactSlice(val)
		default:
			result[i] = v
		}
	}
	return result
}

// redactStruct 脱敏结构体
func redactStruct(v interface{}) interface{} {
	// 尝试将结构体转换为 map
	data, err := json.Marshal(v)
	if err != nil {
		return v
	}
	
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return v
	}
	
	return redactMap(m)
}

// RedactField 脱敏 zap.Field
func RedactField(field zapcore.Field) zapcore.Field {
	// 检查字段名是否敏感
	if isSensitiveField(field.Key) {
		field.String = RedactedValue
		field.Interface = RedactedValue
		return field
	}
	
	// 对于 ReflectType 类型，需要深度脱敏
	if field.Type == zapcore.ReflectType && field.Interface != nil {
		val := reflect.ValueOf(field.Interface)
		
		// 处理指针
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return field
			}
			val = val.Elem()
		}
		
		switch val.Kind() {
		case reflect.Map:
			// 尝试转换为 map[string]interface{}
			if data, err := json.Marshal(field.Interface); err == nil {
				var m map[string]interface{}
				if err := json.Unmarshal(data, &m); err == nil {
					field.Interface = redactMap(m)
				}
			}
		case reflect.Struct:
			field.Interface = redactStruct(field.Interface)
		case reflect.Slice, reflect.Array:
			if data, err := json.Marshal(field.Interface); err == nil {
				var s []interface{}
				if err := json.Unmarshal(data, &s); err == nil {
					field.Interface = redactSlice(s)
				}
			}
		}
	}
	
	return field
}

// RedactingCore 实现脱敏的 zapcore.Core
type RedactingCore struct {
	zapcore.Core
}

// NewRedactingCore 创建脱敏 Core
func NewRedactingCore(core zapcore.Core) zapcore.Core {
	return &RedactingCore{Core: core}
}

// With 添加字段时进行脱敏
func (c *RedactingCore) With(fields []zapcore.Field) zapcore.Core {
	redactedFields := make([]zapcore.Field, len(fields))
	for i, field := range fields {
		redactedFields[i] = RedactField(field)
	}
	return &RedactingCore{Core: c.Core.With(redactedFields)}
}

// Check 检查是否需要记录
func (c *RedactingCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

// Write 写入日志时进行脱敏
func (c *RedactingCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	redactedFields := make([]zapcore.Field, len(fields))
	for i, field := range fields {
		redactedFields[i] = RedactField(field)
	}
	return c.Core.Write(entry, redactedFields)
}

// RedactString 脱敏字符串（用于非结构化日志）
func RedactString(s string) string {
	// 脱敏可能包含敏感信息的模式
	patterns := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		{
			// 脱敏 URL 中的密码: user:password@host -> user:***REDACTED***@host
			regex:       regexp.MustCompile(`://([^:@]+):([^@]+)@`),
			replacement: "://$1:" + RedactedValue + "@",
		},
		{
			regex:       regexp.MustCompile(`(?i)(password|passwd|pwd)[\s]*[:=][\s]*[^\s&,;]+`),
			replacement: "$1=" + RedactedValue,
		},
		{
			regex:       regexp.MustCompile(`(?i)(secret|api[_-]?key|token)[\s]*[:=][\s]*[^\s&,;]+`),
			replacement: "$1=" + RedactedValue,
		},
		{
			regex:       regexp.MustCompile(`(?i)(authorization|auth)[\s]*:[\s]*[^\s,;]+`),
			replacement: "$1: " + RedactedValue,
		},
		{
			regex:       regexp.MustCompile(`(?i)(private[_-]?key)[\s]*[:=][\s]*[^\s&,;]+`),
			replacement: "$1=" + RedactedValue,
		},
	}
	
	result := s
	for _, p := range patterns {
		result = p.regex.ReplaceAllString(result, p.replacement)
	}
	
	return result
}

