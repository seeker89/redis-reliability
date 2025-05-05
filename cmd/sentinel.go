package cmd

import (
	"context"

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
	sentinelCmd.PersistentFlags().StringVar(&redisCfg.SentinelURL, "sentinel", "redis://redis.cluster.local:26379", "Redis URL of the sentinel")
	sentinelCmd.PersistentFlags().StringVar(&redisCfg.SentinelMaster, "master", "mymaster", "Redis master name")
}
