package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func writer(url, key string, values chan string, n int) {
	rdb := redis.NewClient(&redis.Options{Addr: url})

	for i := range n {
		val := fmt.Sprint(i)
		_, err := rdb.Set(ctx, key, val, 0).Result()
		if err != nil {
			panic(err)
		}
		// send the value and wait until reader finished
		values <- val
		<-values
	}
	values <- "done"
}

func reader(url, key string, values chan string) {
	rdb := redis.NewClient(&redis.Options{Addr: url})

	var expected_val string
	stales := 0
	total := 0
	for {
		expected_val = <-values
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			panic(err)
		}
		if expected_val == "done" {
			error_rate := float32(stales) / float32(total)
			fmt.Println(
				"Done:", key,
				"total_reads:", total,
				"stale_reads:", stales,
				"error_rate:", error_rate,
			)
			break
		}
		total += 1
		if val != expected_val {
			stales += 1
			fmt.Println(
				"Wrong value:", key,
				"got:", val,
				"expected:", expected_val,
			)
		}
		values <- ""
	}
}

func main() {
	url_write := strings.TrimPrefix(os.Getenv("URL_M"), "redis://")
	url_read := strings.TrimPrefix(os.Getenv("URL_R"), "redis://")

	var wg sync.WaitGroup

	clients := 100
	operations := 1000

	for i := range clients {
		wg.Add(2)
		key := fmt.Sprintf("client_%d", i)
		values := make(chan string)
		go func() {
			defer wg.Done()
			writer(url_write, key, values, operations)
		}()
		go func() {
			defer wg.Done()
			reader(url_read, key, values)
		}()
	}
	wg.Wait()
	fmt.Println(clients, "clients all done")
}
