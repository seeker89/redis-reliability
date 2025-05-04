package config

import "github.com/redis/go-redis/v9"

type RedisConfig struct {
	SentinelURL    string
	SentinelMaster string
	RedisClient    *redis.Client
}
