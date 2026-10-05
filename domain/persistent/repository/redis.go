package repository

import (
	"context"
	"time"
)

type RedisRepo interface {
	Get(ctx context.Context, key string) (string, error)
	GetBytes(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNx(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	NewLock(key string, expiration time.Duration) DistributedLock
	Del(ctx context.Context, key string) bool
	// MSet 批量设置 key-value，keys 与 values 一一对应，无过期时间
	MSet(ctx context.Context, keys []string, values []string) error
	// SetManyWithExpiry 批量设置 key-value 并统一过期时间，用 Pipeline 执行，keys 与 values 一一对应
	SetManyWithExpiry(ctx context.Context, keys []string, values []string, expiration time.Duration) error
	// MGet 批量获取，返回与 keys 顺序一致；不存在的 key 对应空字符串
	MGet(ctx context.Context, keys ...string) ([]string, error)
}

type DistributedLock interface {
	Lock(ctx context.Context) bool
	TryLock(ctx context.Context) bool
	Unlock(ctx context.Context) bool
}
