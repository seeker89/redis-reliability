package cmd

import (
	"fmt"
	"os"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current master of the cluster",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteSentinelStatus(&cfg, &redisCfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(statusCmd)
}

func ExecuteSentinelStatus(
	config *config.RRTConfig,
	redisConfig *config.RedisSentinelConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	master, err := redisClient.GetMasterFromSentinel(ctx, rdb, redisConfig.SentinelMaster)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	printer.Itemise = true
	printer.Print([]map[string]string{
		{
			"host": master.Host,
			"port": master.Port,
		},
	}, []string{})
	return nil
}
