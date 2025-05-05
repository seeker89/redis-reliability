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

var sentinelMasterCmd = &cobra.Command{
	Use:   "master",
	Short: "Show the details of the redis master",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteSentinelMasters(&cfg, &redisCfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelMasterCmd)
}

func ExecuteSentinelMasters(
	config *config.RRTConfig,
	redisConfig *config.RedisConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	{
		cmd := redis.NewMapStringStringCmd(ctx, "SENTINEL", "master", redisConfig.SentinelMaster)
		if err := rdb.Process(ctx, cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		res, _ := cmd.Result()
		printer.Print(
			[]map[string]string{res},
			[]string{
				"name",
				"quorum",
				"config-epoch",
				"num-slaves",
				"port",
				"ip",
			},
		)
	}
	return nil
}
