package cmd

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the sentinel cluster",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteSentinelStatus(&cfg, &redisCfg)
	},
}

func init() {
	sentinelCmd.AddCommand(statusCmd)
}

func ExecuteSentinelStatus(config *config.RRTConfig, redisConfig *config.RedisConfig) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	fmt.Println("MASTER ADDR")
	cmd := rdb.Do(ctx, "SENTINEL", "get-master-addr-by-name", redisConfig.SentinelMaster)
	rdb.Process(ctx, cmd)
	res, err := cmd.StringSlice()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(res)
	{
		cmd := redis.NewMapStringStringCmd(ctx, "SENTINEL", "master", redisConfig.SentinelMaster)
		rdb.Process(ctx, cmd)
		res, _ := cmd.Result()
		fmt.Println("MASTER CFG")
		fmt.Println(res)
	}
	{
		//cmd := rdb.Do(ctx, "SENTINEL", "sentinels", redisConfig.SentinelMaster)
		cmd := redis.NewMapStringStringSliceCmd(ctx, "SENTINEL", "sentinels", redisConfig.SentinelMaster)
		rdb.Process(ctx, cmd)
		res, _ := cmd.Result()
		fmt.Println("SENTINEL CONFIG")
		fmt.Println(res)
	}
	return nil
}
