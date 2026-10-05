package redis

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/redis"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	goredislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type Client struct {
	rdb goredislib.UniversalClient
	rs  *redsync.Redsync
}

func NewClient(conf *conf.Config) repository.RedisRepo {
	rdb := redis.NewRedisClient(conf.RedisConf)
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	return &Client{
		rdb: rdb,
		rs:  rs,
	}
}

func (c *Client) Get(ctx context.Context, realKey string) (string, error) {
	val, err := c.rdb.Get(ctx, realKey).Result()
	if err == goredislib.Nil {
		logger.InfoZ(ctx, "key does not exist", zap.String("key", realKey))
	} else if err != nil {
		notification.SendErrorLog(ctx, "redis get error", zap.String("key", realKey), zap.String("error", err.Error()))
	}
	return val, err
}

func (c *Client) Incr(ctx context.Context, realKey string) (int64, error) {
	val, err := c.rdb.Incr(ctx, realKey).Result()
	if err == goredislib.Nil {
		logger.InfoZ(ctx, "key does not exist", zap.String("key", realKey))
	} else if err != nil {
		notification.SendErrorLog(ctx, "redis get error", zap.String("key", realKey), zap.String("error", err.Error()))
	}
	return val, err
}

func (c *Client) GetBytes(ctx context.Context, realKey string) ([]byte, error) {
	val, err := c.rdb.Get(ctx, realKey).Bytes()
	if err == goredislib.Nil {
		logger.InfoZ(ctx, "key does not exist", zap.String("key", realKey))
	} else if err != nil {
		notification.SendErrorLog(ctx, "redis get error", zap.String("key", realKey), zap.String("error", err.Error()))
	}
	return val, err
}

func (c *Client) Set(ctx context.Context, realKey string, value interface{}, expiration time.Duration) error {
	err := c.rdb.Set(ctx, realKey, value, expiration).Err()
	if err != nil {
		notification.SendErrorLog(ctx, "redis set error", zap.String("key", realKey), zap.String("error", err.Error()))
	}
	return err
}

func (c *Client) SetNx(ctx context.Context, realKey string, value interface{}, expiration time.Duration) (bool, error) {
	ok, err := c.rdb.SetNX(ctx, realKey, value, expiration).Result()
	if err != nil {
		notification.SendErrorLog(ctx, "redis SetNx error", zap.String("key", realKey), zap.String("error", err.Error()))
		return false, err
	}

	return ok, err
}

func (c *Client) Del(ctx context.Context, realKey string) bool {
	delCount, err := c.rdb.Del(ctx, realKey).Result()
	if err != nil {
		notification.SendErrorLog(ctx, "redis SetNx error", zap.String("key", realKey), zap.String("error", err.Error()))
		return false
	}

	if delCount <= 0 {
		return false
	}

	return true
}

func (c *Client) MSet(ctx context.Context, keys []string, values []string) error {
	if len(keys) != len(values) || len(keys) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(keys)+len(values))
	for i := range keys {
		args = append(args, keys[i], values[i])
	}
	err := c.rdb.MSet(ctx, args...).Err()
	if err != nil {
		notification.SendErrorLog(ctx, "redis MSet error", zap.Int("count", len(keys)), zap.String("error", err.Error()))
	}
	return err
}

func (c *Client) SetManyWithExpiry(ctx context.Context, keys []string, values []string, expiration time.Duration) error {
	if len(keys) != len(values) || len(keys) == 0 {
		return nil
	}
	pipe := c.rdb.Pipeline()
	for i := range keys {
		pipe.Set(ctx, keys[i], values[i], expiration)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		notification.SendErrorLog(ctx, "redis SetManyWithExpiry error", zap.Int("count", len(keys)), zap.String("error", err.Error()))
	}
	return err
}

func (c *Client) MGet(ctx context.Context, keys ...string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	vals, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		notification.SendErrorLog(ctx, "redis MGet error", zap.Int("count", len(keys)), zap.String("error", err.Error()))
		return nil, err
	}
	out := make([]string, len(vals))
	for i, v := range vals {
		if v == nil {
			continue
		}
		switch s := v.(type) {
		case string:
			out[i] = s
		case []byte:
			out[i] = string(s)
		default:
			out[i] = fmt.Sprintf("%v", v)
		}
	}
	return out, nil
}

func (c *Client) GetKeyPrefix() string {
	return fmt.Sprintf("%s:%s:", conf.GetServerName(), conf.GetEnv())
}

func (c *Client) NewLock(realKey string, expiration time.Duration) repository.DistributedLock {

	return &DistributedLock{
		mutex: c.rs.NewMutex(realKey, redsync.WithExpiry(expiration)),
		key:   realKey,
	}
}

type DistributedLock struct {
	key   string
	mutex *redsync.Mutex
}

// Lock
// 这个会循环重试，直到重试的最大次数
func (d *DistributedLock) Lock(ctx context.Context) bool {
	if err := d.mutex.LockContext(ctx); err != nil {
		notification.SendErrorLog(ctx, "redis lock error", zap.String("error", err.Error()), zap.String("key", d.key))
		return false
	}
	return true
}

func (d *DistributedLock) TryLock(ctx context.Context) bool {
	if err := d.mutex.TryLockContext(ctx); err != nil {
		// 如果错误是 ErrFailed，说明锁已被占用，这是正常情况，不需要报警
		// 其他错误（如连接错误）也不报警，避免频繁报警
		// if err != redsync.ErrFailed {
		// 	notification.SendErrorLog(ctx, "redis try lock error", zap.String("error", err.Error()), zap.String("key", d.key))
		// }
		return false
	}
	return true
}

func (d *DistributedLock) Unlock(ctx context.Context) bool {
	ok, err := d.mutex.UnlockContext(ctx)
	if err != nil {
		// 解锁错误（如网络错误、Redis 连接错误）是常见情况，不影响业务逻辑
		// 因为锁已经过期或业务已完成，即使解锁失败也不影响数据一致性
		// 只记录日志，不报警，避免频繁报警
		logger.WarnZ(ctx, "redis unlock error",
			zap.String("error", err.Error()),
			zap.String("key", d.key),
			zap.String("reason", "解锁失败，但锁可能已过期，不影响业务逻辑"))
		return false
	}
	if !ok {
		// ok 为 false 表示锁已过期或 token 不匹配，这是正常情况（锁可能已被其他协程获取或已过期）
		// 只记录日志，不报警，避免频繁报警
		logger.InfoZ(ctx, "redis unlock failed (lock expired or token mismatch)",
			zap.String("key", d.key),
			zap.String("reason", "锁已过期或 token 不匹配，这是正常情况"))
		return false
	}
	return true
}
