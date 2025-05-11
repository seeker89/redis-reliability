package redisClient

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisInstance struct {
	Host string
	Port string
}

func MakeRedisClient(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opts), nil
}

func GetMasterFromSentinel(rdb *redis.Client, ctx context.Context, master string) (*RedisInstance, error) {
	cmd := rdb.Do(ctx, "SENTINEL", "get-master-addr-by-name", master)
	if err := rdb.Process(ctx, cmd); err != nil {
		return nil, err
	}
	res, err := cmd.StringSlice()
	if err != nil {
		return nil, err
	}
	return &RedisInstance{
		Host: res[0],
		Port: res[1],
	}, nil
}
