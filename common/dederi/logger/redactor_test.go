package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		expected  bool
	}{
		{"password exact match", "password", true},
		{"PASSWORD uppercase", "PASSWORD", true},
		{"api_key exact match", "api_key", true},
		{"apiKey camelCase", "apiKey", true},
		{"privateKey", "privateKey", true},
		{"private_key", "private_key", true},
		{"token", "token", true},
		{"secret", "secret", true},
		{"normal field", "user_name", false},
		{"normal field", "email", false},
		{"contains password", "user_password", true},
		{"api key variation", "api_key", true},
		{"private key variation", "encryption_private_key", true},
		{"access key", "aws_access_key", true},
		// 不应该被脱敏的字段
		{"normal_key should not match", "normal_key", false},
		{"public_key should not match", "public_key", false},
		{"user_id", "user_id", false},
		{"market_id", "market_id", false},
		{"key_name (安全)", "key_name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveField(tt.fieldName)
			if result != tt.expected {
				t.Errorf("isSensitiveField(%s) = %v, expected %v", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestRedactMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "simple password redaction",
			input: map[string]interface{}{
				"username": "admin",
				"password": "secret123",
			},
			expected: map[string]interface{}{
				"username": "admin",
				"password": RedactedValue,
			},
		},
		{
			name: "nested map redaction",
			input: map[string]interface{}{
				"user": "admin",
				"config": map[string]interface{}{
					"api_key": "abc123",
					"timeout": 30,
				},
			},
			expected: map[string]interface{}{
				"user": "admin",
				"config": map[string]interface{}{
					"api_key": RedactedValue,
					"timeout": 30,
				},
			},
		},
		{
			name: "multiple sensitive fields",
			input: map[string]interface{}{
				"username":    "admin",
				"password":    "pass123",
				"api_key":     "key123",
				"api_secret":  "secret123",
				"private_key": "pk123",
			},
			expected: map[string]interface{}{
				"username":    "admin",
				"password":    RedactedValue,
				"api_key":     RedactedValue,
				"api_secret":  RedactedValue,
				"private_key": RedactedValue,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactMap(tt.input)
			for k, expectedV := range tt.expected {
				resultV, exists := result[k]
				if !exists {
					t.Errorf("Key %s not found in result", k)
					continue
				}

				// 对于嵌套的 map，递归比较
				if expectedMap, ok := expectedV.(map[string]interface{}); ok {
					resultMap, ok := resultV.(map[string]interface{})
					if !ok {
						t.Errorf("Value for key %s is not a map", k)
						continue
					}
					for nestedK, nestedExpectedV := range expectedMap {
						if resultMap[nestedK] != nestedExpectedV {
							t.Errorf("Nested key %s.%s = %v, expected %v", k, nestedK, resultMap[nestedK], nestedExpectedV)
						}
					}
				} else if resultV != expectedV {
					t.Errorf("Key %s = %v, expected %v", k, resultV, expectedV)
				}
			}
		})
	}
}

func TestRedactField(t *testing.T) {
	tests := []struct {
		name     string
		field    zapcore.Field
		expected string
	}{
		{
			name:     "password field",
			field:    zap.String("password", "secret123"),
			expected: RedactedValue,
		},
		{
			name:     "api_key field",
			field:    zap.String("api_key", "key123"),
			expected: RedactedValue,
		},
		{
			name:     "normal field",
			field:    zap.String("username", "admin"),
			expected: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactField(tt.field)
			if result.String != tt.expected {
				t.Errorf("RedactField(%s) = %s, expected %s", tt.field.Key, result.String, tt.expected)
			}
		})
	}
}

func TestRedactString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // 检查结果是否包含此字符串
	}{
		{
			name:     "password in URL",
			input:    "database://user:password123@localhost",
			contains: RedactedValue,
		},
		{
			name:     "api_key parameter",
			input:    "api_key=abc123&user=admin",
			contains: RedactedValue,
		},
		{
			name:     "authorization header",
			input:    "Authorization: Bearer token123",
			contains: RedactedValue,
		},
		{
			name:     "normal string",
			input:    "This is a normal log message",
			contains: "normal log message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactString(tt.input)
			if tt.contains == RedactedValue {
				// 检查是否包含脱敏值
				if result == tt.input {
					t.Errorf("RedactString should have redacted sensitive info in: %s", tt.input)
				}
			}
		})
	}
}

