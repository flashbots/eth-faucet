package ratelimiter

import (
	"context"
	"errors"
	"time"

	"github.com/flashbots/eth-faucet/backoff"
	"github.com/flashbots/eth-faucet/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimiter struct {
	backoffParams *backoff.Parameters
	prefix        string
	redis         *redis.Client
}

func New(cfg *config.Config) (*RateLimiter, error) {
	l := zap.L()

	backoffParams := &backoff.Parameters{
		BaseTimeout: cfg.Redis.Timeout,
	}

	redisOptions, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, err
	}
	_redis := redis.NewClient(redisOptions)

	l.Info("Connecting to redis...", zap.String("redis_url", cfg.Redis.URL))
	err = backoff.Backoff(context.Background(), backoffParams, func(ctx context.Context) (_err error) {
		status := _redis.Ping(ctx)
		_err = status.Err()
		if _err != nil {
			l.Warn("Failed to connect to redis", zap.Error(_err))
		}
		return
	})
	if err != nil {
		return nil, err
	}

	prefix := ""
	if cfg.Redis.Namespace != "" {
		prefix = cfg.Redis.Namespace + ":"
	}

	return &RateLimiter{
		backoffParams: backoffParams,
		prefix:        prefix,
		redis:         _redis,
	}, nil
}

func (rl *RateLimiter) Register(ctx context.Context, key string, expiration time.Duration) error {
	key = rl.prefix + key
	return backoff.Backoff(ctx, rl.backoffParams, func(ctx context.Context) (_err error) {
		_, _err = rl.redis.Set(
			ctx,
			key,
			time.Now().Format(time.RFC3339),
			expiration,
		).Result()
		return
	})
}

func (rl *RateLimiter) IsRegistered(ctx context.Context, key string) (time.Time, error) {
	key = rl.prefix + key

	var res string
	err := backoff.Backoff(ctx, rl.backoffParams, func(ctx context.Context) (_err error) {
		res, _err = rl.redis.Get(
			ctx,
			key,
		).Result()
		return
	})

	redisNil := redis.Nil
	if errors.As(err, &redisNil) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, res)
}
