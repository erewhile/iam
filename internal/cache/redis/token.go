package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/erewhile/iam/internal/consts"
	"github.com/google/uuid"
)

type TokenCache interface {
	SetAccess(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error
	ExistsAccess(ctx context.Context, sessionID uuid.UUID) (bool, error)
	DelAccess(ctx context.Context, sessionID uuid.UUID) error
	SetRefresh(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error
	ExistsRefresh(ctx context.Context, sessionID uuid.UUID) (bool, error)
	DelRefresh(ctx context.Context, sessionID uuid.UUID) error
}

type tokenCache struct {
	rdb *Redis
}

func NewTokenCache() TokenCache {
	return &tokenCache{rdb: DB()}
}

func (t *tokenCache) accessKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("%s:%s", consts.RedisAccessTokenKey, sessionID.String())
}

func (t *tokenCache) SetAccess(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error {
	return t.rdb.Set(ctx, t.accessKey(sessionID), "1", ttl)
}

func (t *tokenCache) ExistsAccess(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return t.rdb.Exists(ctx, t.accessKey(sessionID))
}

func (t *tokenCache) DelAccess(ctx context.Context, sessionID uuid.UUID) error {
	return t.rdb.Del(ctx, t.accessKey(sessionID))
}

func (t *tokenCache) refreshKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("%s:%s", consts.RedisRefreshTokenKey, sessionID.String())
}

func (t *tokenCache) SetRefresh(ctx context.Context, sessionID uuid.UUID, ttl time.Duration) error {
	return t.rdb.Set(ctx, t.refreshKey(sessionID), "1", ttl)
}

func (t *tokenCache) ExistsRefresh(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	return t.rdb.Exists(ctx, t.refreshKey(sessionID))
}

func (t *tokenCache) DelRefresh(ctx context.Context, sessionID uuid.UUID) error {
	return t.rdb.Del(ctx, t.refreshKey(sessionID))
}
