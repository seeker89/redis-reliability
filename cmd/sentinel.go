package cmd

import (
	"context"
	"os"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/spf13/cobra"
)

var redisCfg config.RedisConfig
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
}
