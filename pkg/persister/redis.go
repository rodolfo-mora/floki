package persister

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type RedisPersister struct {
	ctx context.Context
	cli *redis.Client
}

// Returns an instance of RedisPersister interface.
//
//	ctx context.Context
//	cli *redis.Client
func NewRedisPersister(ctx context.Context, cli *redis.Client) RedisPersister {
	return RedisPersister{
		ctx: ctx,
		cli: cli,
	}
}

func (r RedisPersister) Set(key string, value interface{}) error {
	return insert(r.ctx, key, value, r.cli)
}

func (r RedisPersister) Get(key string) (string, error) {
	return get(r.ctx, key, r.cli)
}

func (r RedisPersister) Delete(key string) error {
	return delete(r.ctx, key, r.cli)
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:       "localhost:6379",
		Password:   "", // no password set
		DB:         0,  // use default DB
		MaxRetries: 3,
	})
}

func insert(ctx context.Context, key string, value interface{}, rdb *redis.Client) error {
	var err error
	var v []byte

	v, err = marshall(value)
	if err != nil {
		return err
	}

	return (*rdb).Set(ctx, key, string(v), 0).Err()
}

func delete(ctx context.Context, key string, rdb *redis.Client) error {
	return rdb.Del(ctx, key).Err()
}

func get(ctx context.Context, name string, rdb *redis.Client) (string, error) {
	return rdb.Get(ctx, name).Result()
}

func marshall(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}
