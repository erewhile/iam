package rds

import (
	"context"
	"encoding/json"
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
	SetCode(ctx context.Context, code string, payload OAuthCodePayload) error
	GetAndDelCode(ctx context.Context, code string) (*OAuthCodePayload, error)
}

type OAuthCodePayload struct {
	UserID    int       `json:"user_id"`
	UserUUID  uuid.UUID `json:"user_uuid"`
	SessionID uuid.UUID `json:"session_id"`
	ClientID  string    `json:"client_id"`
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

func (t *tokenCache) SetCode(ctx context.Context, code string, payload OAuthCodePayload) error {
	return t.rdb.SetJSON(ctx, fmt.Sprintf("%s:%s", consts.RedisOAuthCodeKey, code), payload, 5*time.Minute)
}

func (t *tokenCache) GetAndDelCode(ctx context.Context, code string) (*OAuthCodePayload, error) {
	key := fmt.Sprintf("%s:%s", consts.RedisOAuthCodeKey, code)

	luaScript := `
		local val = redis.call("GET", KEYS[1])
		if val then
			redis.call("DEL", KEYS[1])
		end
		return val
	`

	res, err := t.rdb.Client().Eval(ctx, luaScript, []string{t.rdb.key(key)}, nil).Result()
	if err != nil {
		return nil, err
	}

	valStr, ok := res.(string)
	if !ok || valStr == "" {
		return nil, fmt.Errorf("code not found or expired")
	}

	var payload OAuthCodePayload
	if err := json.Unmarshal([]byte(valStr), &payload); err != nil {
		return nil, fmt.Errorf("unmarshal oauth payload failed: %w", err)
	}

	return &payload, nil
}
