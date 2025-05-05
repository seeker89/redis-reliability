package cmd

import (
	"github.com/redis/go-redis/v9"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var sentinelSentinelsCmd = &cobra.Command{
	Use:   "sentinels",
	Short: "Show the sentinels for a master",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteSentinelSentinels(&cfg, &redisCfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelSentinelsCmd)
}

func ExecuteSentinelSentinels(
	config *config.RRTConfig,
	redisConfig *config.RedisConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	{
		cmd := redis.NewMapStringStringSliceCmd(ctx, "SENTINEL", "sentinels", redisConfig.SentinelMaster)
		rdb.Process(ctx, cmd)
		res, _ := cmd.Result()
		printer.Print(
			res,
			[]string{
				"name",
				"voted-leader",
				"voted-leader-epoch",
				"port",
				"ip",
			},
		)
	}
	return nil
}
