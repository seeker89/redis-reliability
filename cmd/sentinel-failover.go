package cmd

import (
	"fmt"
	"os"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var sentinelFailoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "Trigger soft redis failover",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ExecuteSentinelFailover(&cfg, &redisCfg, prtr); err == nil {
			ExecuteSentinelStatus(&cfg, &redisCfg, prtr)
		}
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelFailoverCmd)
}

func ExecuteSentinelFailover(
	config *config.RRTConfig,
	redisConfig *config.RedisConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	{
		cmd := rdb.Do(ctx, "SENTINEL", "failover", redisConfig.SentinelMaster)
		if err := rdb.Process(ctx, cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		res, err := cmd.Text()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		printer.Print([]map[string]string{
			{
				"result": res,
			},
		}, []string{})
	}
	return nil
}
