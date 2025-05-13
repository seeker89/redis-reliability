package cmd

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var sentinelSentinelsCmd = &cobra.Command{
	Use:   "sentinels",
	Short: "Show the sentinels for a master",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteSentinelSentinels(&cfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelSentinelsCmd)
}

func ExecuteSentinelSentinels(
	config *config.RRConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(config.SentinelURL)
	if err != nil {
		return err
	}
	{
		cmd := redis.NewMapStringStringSliceCmd(ctx, "SENTINEL", "sentinels", config.SentinelMaster)
		if err := rdb.Process(ctx, cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
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
