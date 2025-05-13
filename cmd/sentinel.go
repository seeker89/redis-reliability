package cmd

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
)

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
		&cfg.SentinelURL,
		"sentinel",
		os.Getenv(CMD_PREFIX+"SENTINEL_URL"),
		"Redis URL of the sentinel. Use "+CMD_PREFIX+"SENTINEL_URL",
	)
	sentinelCmd.PersistentFlags().StringVar(&cfg.SentinelMaster, "master", master, "Redis master name")
	sentinelCmd.PersistentFlags().DurationVarP(&cfg.Timeout, "timeout", "t", 60*time.Second, "Timeout for killing")
	sentinelCmd.PersistentFlags().DurationVarP(&cfg.Grace, "grace", "g", 0*time.Second, "Grace period for killing")
}
