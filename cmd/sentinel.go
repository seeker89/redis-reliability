package cmd

import (
	"context"
	"os"
	"time"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/spf13/cobra"
)

var redisCfg config.RedisSentinelConfig
var ctx = context.Background()

var sentinelCmd = &cobra.Command{
	Use:   "sentinel",
	Short: "Verify Redis sentinel setup",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(sentinelCmd)
	master := os.Getenv(CMD_PREFIX + "SENTINEL_MASTER")
	if master == "" {
		master = "mymaster"
	}
	sentinelCmd.PersistentFlags().StringVar(
		&redisCfg.SentinelURL,
		"sentinel",
		os.Getenv(CMD_PREFIX+"SENTINEL_URL"),
		"Redis URL of the sentinel. Use "+CMD_PREFIX+"SENTINEL_URL",
	)
	sentinelCmd.PersistentFlags().StringVar(&redisCfg.SentinelMaster, "master", master, "Redis master name")
	sentinelCmd.PersistentFlags().DurationVarP(&redisCfg.Timeout, "timeout", "t", 60*time.Second, "Timeout for killing")
}
