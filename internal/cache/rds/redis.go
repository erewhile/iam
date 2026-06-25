package rds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/consts"
	"github.com/redis/go-redis/v9"
)

const pingTimeout = 5 * time.Second

type Redis struct {
	client *redis.Client
	prefix string
}

var (
	mu  sync.RWMutex
	rdb *Redis
)

func Init(cfg config.Redis) error {
	opt := &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,
	}

	c := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		_ = c.Close()
		return fmt.Errorf("redis ping failed: %w", err)
	}

	mu.Lock()
	rdb = &Redis{client: c, prefix: cfg.Prefix}
	mu.Unlock()

	return nil
}

func DB() *Redis {
	mu.RLock()
	defer mu.RUnlock()
	if rdb == nil {
		log.Fatal("redis client is not initialized, call Init() first")
	}
	return rdb
}

func (r *Redis) key(k string) string {
	if r.prefix == "" {
		return k
	}
	return r.prefix + ":" + k
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func Close() error {
	mu.Lock()
	r := rdb
	rdb = nil
	mu.Unlock()

	if r == nil || r.client == nil {
		return nil
	}
	if err := r.client.Close(); err != nil {
		log.Printf("error failed to close redis: %v\n", err)
		return err
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	s, err := r.client.Get(ctx, r.key(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", redis.Nil
	}
	if err != nil {
		return "", fmt.Errorf("get key %s: %w", key, err)
	}
	return s, nil
}

func (r *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := r.client.Set(ctx, r.key(key), value, ttl).Err(); err != nil {
		return fmt.Errorf("set key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) SetForever(ctx context.Context, key string, value any) error {
	if err := r.client.Set(ctx, r.key(key), value, 0).Err(); err != nil {
		return fmt.Errorf("set forever key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) SetNX(ctx context.Context, key string, value any, exp time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, r.key(key), value, exp).Result()
	if err != nil {
		return false, fmt.Errorf("setnx key %s: %w", key, err)
	}
	return ok, nil
}

func (r *Redis) Del(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, r.key(key)).Err(); err != nil {
		return fmt.Errorf("del key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, r.key(key)).Result()
	if err != nil {
		return false, fmt.Errorf("exists key %s: %w", key, err)
	}
	return n > 0, nil
}

func (r *Redis) SAdd(ctx context.Context, key string, members ...any) error {
	if err := r.client.SAdd(ctx, r.key(key), members...).Err(); err != nil {
		return fmt.Errorf("sadd key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	sids, err := r.client.SMembers(ctx, r.key(key)).Result()
	if err != nil {
		return nil, fmt.Errorf("smembers key %s: %w", key, err)
	}
	return sids, nil
}

func (r *Redis) SRem(ctx context.Context, key string, members ...any) error {
	if err := r.client.SRem(ctx, r.key(key), members...).Err(); err != nil {
		return fmt.Errorf("srem key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value for key %s: %w", key, err)
	}
	return r.Set(ctx, key, b, ttl)
}

func (r *Redis) GetJSON(ctx context.Context, key string, dst any) error {
	s, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(s), dst); err != nil {
		return fmt.Errorf("unmarshal value for key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, r.key(key), ttl).Err(); err != nil {
		return fmt.Errorf("expire key %s: %w", key, err)
	}
	return nil
}

func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	n, err := r.client.Incr(ctx, r.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("incr key %s: %w", key, err)
	}
	return n, nil
}

func (r *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	n, err := r.client.IncrBy(ctx, r.key(key), value).Result()
	if err != nil {
		return 0, fmt.Errorf("incrby key %s: %w", key, err)
	}
	return n, nil
}

func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	d, err := r.client.TTL(ctx, r.key(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("ttl key %s: %w", key, err)
	}
	return d, nil
}

func (r *Redis) CuckooFilterAdd(ctx context.Context, key, item string) error {
	if err := r.client.Do(ctx, "CF.ADD", r.key(consts.RedisCuckooFilterKeyPrefix+key), item).Err(); err != nil {
		return fmt.Errorf("cf.add key %s item %s: %w", key, item, err)
	}
	return nil
}

func (r *Redis) CuckooFilterExists(ctx context.Context, key, item string) (bool, error) {
	res, err := r.client.Do(ctx, "CF.EXISTS", r.key(consts.RedisCuckooFilterKeyPrefix+key), item).Result()
	if err != nil {
		return false, fmt.Errorf("cf.exists key %s item %s: %w", key, item, err)
	}
	n, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected cf.exists result type for key %s", key)
	}
	return n == 1, nil
}

func (r *Redis) CuckooFilterDel(ctx context.Context, key, item string) error {
	if err := r.client.Do(ctx, "CF.DEL", r.key(consts.RedisCuckooFilterKeyPrefix+key), item).Err(); err != nil {
		return fmt.Errorf("cf.del key %s item %s: %w", key, item, err)
	}
	return nil
}

func (r *Redis) CuckooFilterAddNX(ctx context.Context, key, item string) (bool, error) {
	res, err := r.client.Do(ctx, "CF.ADDNX", r.key(consts.RedisCuckooFilterKeyPrefix+key), item).Result()
	if err != nil {
		return false, fmt.Errorf("cf.addnx key %s item %s: %w", key, item, err)
	}
	n, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected cf.addnx result type for key %s", key)
	}
	return n == 1, nil
}

func (r *Redis) BloomAdd(ctx context.Context, key, item string) error {
	if err := r.client.Do(ctx, "BF.ADD", r.key("bf:"+key), item).Err(); err != nil {
		return fmt.Errorf("bf.add key %s item %s: %w", key, item, err)
	}
	return nil
}

func (r *Redis) BloomExists(ctx context.Context, key, item string) (bool, error) {
	res, err := r.client.Do(ctx, "BF.EXISTS", r.key("bf:"+key), item).Result()
	if err != nil {
		return false, fmt.Errorf("bf.exists key %s item %s: %w", key, item, err)
	}
	n, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected bf.exists result type for key %s", key)
	}
	return n == 1, nil
}
