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

var sentinelReplicasCmd = &cobra.Command{
	Use:   "replicas",
	Short: "Show the details of the replicas for a master",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteSentinelReplicas(&cfg, &redisCfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelReplicasCmd)
}

func ExecuteSentinelReplicas(
	config *config.RRTConfig,
	redisConfig *config.RedisConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	{
		cmd := redis.NewMapStringStringSliceCmd(ctx, "SENTINEL", "replicas", redisConfig.SentinelMaster)
		if err := rdb.Process(ctx, cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		res, _ := cmd.Result()
		printer.Print(
			res,
			[]string{
				"ip",
				"port",
				"slave-repl-offset",
			},
		)
	}
	return nil
}
