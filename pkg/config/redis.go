package config

import "time"

type RedisSentinelConfig struct {
	SentinelURL    string
	SentinelMaster string
	Timeout        time.Duration
}
