package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	config                  *Configuration
	client                  *redis.Client
	decreaseAndDeleteScript *redis.Script
}

func NewRedis(cfg *Configuration) ICache {
	options, err := redis.ParseURL(cfg.Url)
	options.TLSConfig.InsecureSkipVerify = true
	if err != nil {
		log.Fatal(err.Error())
	}
	rdb := redis.NewClient(options)

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err.Error())
	}

	return &redisCache{
		config: cfg,
		client: rdb,
		decreaseAndDeleteScript: redis.NewScript(`
		local key = KEYS[1]

		local value = redis.call('DECR', key)
		if value <= 0 then
			redis.call('DEL', key)
		end
		return value
		`),
	}
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return value, err
	}
	return value, nil
}

func (r *redisCache) SetIfNotExists(ctx context.Context, key string, value interface{}, ttl *time.Duration) (bool, error) {
	var redisTTL time.Duration = redis.KeepTTL
	if ttl != nil {
		redisTTL = *ttl
	}
	return r.client.SetNX(ctx, key, value, redisTTL).Result()
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl *time.Duration) error {
	var redisTTL time.Duration = redis.KeepTTL
	if ttl != nil {
		redisTTL = *ttl
	}
	_, err := r.client.Set(ctx, key, value, redisTTL).Result()
	return err
}

func (r *redisCache) Increase(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisCache) IncreaseBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

func (r *redisCache) Decrease(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

func (r *redisCache) DecreaseAndDelete(ctx context.Context, key string) (int64, error) {
	scriptValue, err := r.decreaseAndDeleteScript.Run(ctx, r.client, []string{key}).Result()
	if err != nil {
		return 0, err
	}
	value, ok := scriptValue.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid value")
	}
	return value, nil
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	_, err := r.client.Del(ctx, key).Result()
	return err
}

func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	value, err := r.client.Exists(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if value == 0 {
		return false, nil
	}
	return true, nil
}
