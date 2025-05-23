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
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteSentinelStatus(&cfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(statusCmd)
}

func ExecuteSentinelStatus(
	config *config.RRConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(cfg.SentinelURL)
	if err != nil {
		return err
	}
	master, err := redisClient.GetMasterFromSentinel(ctx, rdb, cfg.SentinelMaster)
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
