package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/erewhile/iam/internal/consts"
	"github.com/google/uuid"
)

type IAMSessionPayload struct {
	UserID   int       `json:"user_id"`
	UserUUID uuid.UUID `json:"user_uuid"`
}

type IAMSessionCache interface {
	Set(ctx context.Context, sid string, payload IAMSessionPayload, ttl time.Duration) error
	Get(ctx context.Context, sid string) (*IAMSessionPayload, error)
	Del(ctx context.Context, sid string) error
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

func (c *iamSessionCache) Set(ctx context.Context, sid string, payload IAMSessionPayload, ttl time.Duration) error {
	return c.rdb.SetJSON(ctx, c.key(sid), payload, ttl)
}

func (c *iamSessionCache) Get(ctx context.Context, sid string) (*IAMSessionPayload, error) {
	var payload IAMSessionPayload
	err := c.rdb.GetJSON(ctx, c.key(sid), &payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *iamSessionCache) Del(ctx context.Context, sid string) error {
	return c.rdb.Del(ctx, c.key(sid))
}
