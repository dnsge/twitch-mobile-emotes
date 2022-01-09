package storage

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
)

type RedisSettingsRepository struct {
	namespace string
	client    *redis.Client
	ctx       context.Context
}

func NewRedisSettingsRepository(namespace string, options *redis.Options, ctx context.Context) *RedisSettingsRepository {
	client := redis.NewClient(options)

	return &RedisSettingsRepository{
		namespace: namespace,
		client:    client,
		ctx:       ctx,
	}
}

func (r *RedisSettingsRepository) key(name string) string {
	return r.namespace + ":" + name
}

func (r *RedisSettingsRepository) Load(userID string) (*Settings, error) {
	data, err := r.client.Get(r.ctx, r.key("settings:user_id:"+userID)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var s Settings
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *RedisSettingsRepository) Save(userID string, settings *Settings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, r.key("settings:user_id:"+userID), data, 0).Err()
}

func (r *RedisSettingsRepository) Ping() error {
	return r.client.Ping(r.ctx).Err()
}
