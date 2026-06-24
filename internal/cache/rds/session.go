package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type IAMSessionPayload struct {
	UserID   int       `json:"user_id"`
	UserUUID uuid.UUID `json:"user_uuid"`
}

type IAMSessionCache interface {
	Set(ctx context.Context, sid string, payload IAMSessionPayload, ttl time.Duration) error
	Get(ctx context.Context, sid string) (*IAMSessionPayload, error)
	Del(ctx context.Context, sid string) error
	DelAllByUser(ctx context.Context, userID int) error
}

type iamSessionCache struct {
	rdb *Redis
}

func NewIAMSessionCache() IAMSessionCache {
	return &iamSessionCache{rdb: DB()}
}

func (c *iamSessionCache) key(sid string) string {
	return fmt.Sprintf("%s:%s", consts.RedisIAMSessionKey, sid)
}

func (c *iamSessionCache) userSidsKey(userID int) string {
	return fmt.Sprintf("%s:user:%d", consts.RedisIAMSessionKey, userID)
}

func (c *iamSessionCache) Get(ctx context.Context, sid string) (*IAMSessionPayload, error) {
	var payload IAMSessionPayload
	err := c.rdb.GetJSON(ctx, c.key(sid), &payload)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return &payload, nil
}

func (c *iamSessionCache) Set(ctx context.Context, sid string, payload IAMSessionPayload, ttl time.Duration) error {

	if err := c.rdb.SetJSON(ctx, c.key(sid), payload, ttl); err != nil {
		return err
	}

	userSidsKey := c.userSidsKey(payload.UserID)
	if err := c.rdb.SAdd(ctx, userSidsKey, sid); err != nil {
		return err
	}

	return c.rdb.Expire(ctx, userSidsKey, ttl)
}

func (c *iamSessionCache) Del(ctx context.Context, sid string) error {
	payload, err := c.Get(ctx, sid)
	if err == nil && payload != nil {
		userSidsKey := c.userSidsKey(payload.UserID)
		if err := c.rdb.SRem(ctx, userSidsKey, sid); err != nil {
			logger.Error("remove sid from user set failed", err)
		}
	}

	return c.rdb.Del(ctx, c.key(sid))
}

func (c *iamSessionCache) DelAllByUser(ctx context.Context, userID int) error {
	userSidsKey := c.userSidsKey(userID)
	sids, err := c.rdb.SMembers(ctx, userSidsKey)
	if err != nil {
		return err
	}

	for _, sid := range sids {
		if err := c.rdb.Del(ctx, c.key(sid)); err != nil {
			logger.Error("del iam session failed", err)
		}
	}

	return c.rdb.Del(ctx, userSidsKey)
}
