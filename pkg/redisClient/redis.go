package redisClient

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

type RedisInstance struct {
	Host string
	Port string
}

type RedisSwitchMasterEvent struct {
	Master        string
	NewMasterHost string
	NewMasterPort string
	OldMasterHost string
	OldMasterPort string
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

func ParseSwitchMasterMessage(msg string) (*RedisSwitchMasterEvent, error) {
	parts := strings.Split(msg, " ")
	if len(parts) != 5 {
		return nil, fmt.Errorf("expected formatted redis +switch-master, got %s", msg)
	}
	return &RedisSwitchMasterEvent{
		Master:        parts[0],
		OldMasterHost: parts[1],
		OldMasterPort: parts[2],
		NewMasterHost: parts[3],
		NewMasterPort: parts[4],
	}, nil
}
