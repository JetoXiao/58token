package repository

import (
	"crypto/tls"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"

	"github.com/redis/go-redis/v9"
)

// InitRedis initializes the primary business Redis client.
func InitRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(buildRedisOptions(cfg.Redis))
}

// InitResponseCacheRedis initializes the optional AI ResponseCache Redis client.
func InitResponseCacheRedis(cfg *config.Config) *redis.Client {
	if cfg == nil || (!cfg.ResponseCache.Enabled && !cfg.ResponseCache.ShadowEnabled && !cfg.ResponseCache.RecommendationEnabled) {
		return nil
	}
	return redis.NewClient(buildRedisOptions(cfg.ResponseCache.Redis))
}

func buildRedisOptions(redisCfg config.RedisConfig) *redis.Options {
	opts := &redis.Options{
		Addr:         redisCfg.Address(),
		Password:     redisCfg.Password,
		DB:           redisCfg.DB,
		DialTimeout:  time.Duration(redisCfg.DialTimeoutSeconds) * time.Second,
		ReadTimeout:  time.Duration(redisCfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(redisCfg.WriteTimeoutSeconds) * time.Second,
		PoolSize:     redisCfg.PoolSize,
		MinIdleConns: redisCfg.MinIdleConns,
	}

	if redisCfg.EnableTLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: redisCfg.Host,
		}
	}

	return opts
}
