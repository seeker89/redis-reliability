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
	cmd := rdb.Do(ctx, "SENTINEL", "get-master-addr-by-name", redisConfig.SentinelMaster)
	rdb.Process(ctx, cmd)
	res, err := cmd.StringSlice()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	printer.Print([]map[string]string{
		{
			"host": res[0],
			"port": res[1],
		},
	})
	return nil
}
