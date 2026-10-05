package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/redis/go-redis/v9"
)

type RedissonClient struct {
	Mode   string
	Client redis.UniversalClient
}

func NewRedisClient(setting *conf.Redis) redis.UniversalClient {
	if setting == nil {
		panic("redis init error: nil config")
	}
	if setting.PoolSize == 0 {
		setting.PoolSize = 10
	}
	if setting.MinIdle == 0 {
		setting.MinIdle = 5
	}

	addrs := redisAddrs(setting)
	if len(addrs) == 0 {
		panic("redis init error: empty addrs")
	}

	mode := strings.ToLower(strings.TrimSpace(setting.Mode))
	var universalClient redis.UniversalClient

	switch mode {
	case "cluster":
		opt := &redis.ClusterOptions{
			Addrs:        addrs,
			Username:     setting.UserName,
			Password:     setting.Password,
			PoolSize:     setting.PoolSize,
			MinIdleConns: setting.MinIdle,
		}
		if setting.TLS {
			opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		universalClient = redis.NewClusterClient(opt)
	default:
		opt := &redis.UniversalOptions{
			Addrs:        addrs,
			Username:     setting.UserName,
			Password:     setting.Password,
			DB:           setting.DB,
			PoolSize:     setting.PoolSize,
			MinIdleConns: setting.MinIdle,
		}
		if setting.TLS {
			opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		universalClient = redis.NewUniversalClient(opt)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, pingErr := universalClient.Ping(ctx).Result(); pingErr != nil {
		panic(fmt.Sprintf("redis init error (mode=%q addrs=%v): %v", modeOrDefault(mode), addrs, pingErr))
	}

	return universalClient
}

func redisAddrs(setting *conf.Redis) []string {
	addrs := []string{}
	if setting.Port > 0 {
		hosts := strings.Split(setting.Host, ",")
		for _, host := range hosts {
			host = strings.TrimSpace(host)
			if host == "" {
				continue
			}
			addrs = append(addrs, fmt.Sprintf("%s:%d", host, setting.Port))
		}
	} else {
		for _, a := range strings.Split(setting.Host, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				addrs = append(addrs, a)
			}
		}
	}
	return addrs
}

func modeOrDefault(mode string) string {
	if mode == "" {
		return "standalone"
	}
	return mode
}
