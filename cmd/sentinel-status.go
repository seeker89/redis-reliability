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
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteSentinelStatus(&cfg, &redisCfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(statusCmd)
}

func ExecuteSentinelStatus(
	config *config.RRTConfig,
	redisConfig *config.RedisConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	master, err := redisClient.GetMasterFromSentinel(rdb, ctx, redisConfig.SentinelMaster)
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
