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
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteSentinelMasters(&cfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelMasterCmd)
}

func ExecuteSentinelMasters(
	config *config.RRTConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(config.SentinelURL)
	if err != nil {
		return err
	}
	{
		cmd := redis.NewMapStringStringCmd(ctx, "SENTINEL", "master", config.SentinelMaster)
		if err := rdb.Process(ctx, cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		res, _ := cmd.Result()
		printer.Itemise = true
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
