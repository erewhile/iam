package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/erewhile/iam/internal/consts"
)

type LoginAttemptCache interface {
	IncrFailure(ctx context.Context, username string, window time.Duration) (int64, error)
	ResetFailure(ctx context.Context, username string) error
	Lock(ctx context.Context, username string, duration time.Duration) error
	IsLocked(ctx context.Context, username string) (bool, time.Duration, error)
	IncrFailureByIP(ctx context.Context, ip string, window time.Duration) (int64, error)
	ResetFailureByIP(ctx context.Context, ip string) error
}

type loginAttemptCache struct {
	rdb *Redis
}

func NewLoginAttemptCache() LoginAttemptCache {
	return &loginAttemptCache{rdb: DB()}
}

func (c *loginAttemptCache) IncrFailure(ctx context.Context, username string, window time.Duration) (int64, error) {
	key := consts.RedisLoginFailKeyPrefix + username
	count, err := c.rdb.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := c.rdb.Expire(ctx, key, window); err != nil {
			return count, err
		}
	}
	return count, nil
}

func (c *loginAttemptCache) ResetFailure(ctx context.Context, username string) error {
	return c.rdb.Del(ctx, consts.RedisLoginFailKeyPrefix+username)
}

func (c *loginAttemptCache) Lock(ctx context.Context, username string, duration time.Duration) error {
	key := consts.RedisLoginLockKeyPrefix + username
	return c.rdb.Set(ctx, key, "1", duration)
}

func (c *loginAttemptCache) IsLocked(ctx context.Context, username string) (bool, time.Duration, error) {
	key := consts.RedisLoginLockKeyPrefix + username
	ttl, err := c.rdb.TTL(ctx, key)
	if err != nil {
		return false, 0, err
	}
	if ttl <= 0 {
		return false, 0, nil
	}
	return true, ttl, nil
}

func (c *loginAttemptCache) IncrFailureByIP(ctx context.Context, ip string, window time.Duration) (int64, error) {
	key := consts.RedisLoginFailIPKeyPrefix + ip
	count, err := c.rdb.Incr(ctx, key)
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := c.rdb.Expire(ctx, key, window); err != nil {
			return count, err
		}
	}
	return count, nil
}

func (c *loginAttemptCache) ResetFailureByIP(ctx context.Context, ip string) error {
	return c.rdb.Del(ctx, consts.RedisLoginFailIPKeyPrefix+ip)
}

var _ = fmt.Sprintf
