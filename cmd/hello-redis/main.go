package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

func main() {
	// redis-cli likes the protocol specified, go-redis doesn't
	url := strings.TrimPrefix(os.Getenv("URL_M"), "redis://")

	fmt.Println("connecting to redis:", url)
	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var ctx = context.Background()

	val, err := rdb.Get(ctx, "mystery").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("mystery:", val)
}
