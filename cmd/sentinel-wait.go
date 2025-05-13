package cmd

import (
	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var sentinelWaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for the new master election",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteSentinelWait(&cfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelWaitCmd)
}

func ExecuteSentinelWait(
	config *config.RRTConfig,
	printer *printer.Printer,
) error {
	rdb, err := redisClient.MakeRedisClient(cfg.SentinelURL)
	if err != nil {
		return err
	}
	pubsub := rdb.PSubscribe(ctx, "+switch-master")
	defer pubsub.Close()
	printer.Itemise = true
	for msg := range pubsub.Channel() {
		evt, err := redisClient.ParseSwitchMasterMessage(msg.Payload)
		if err != nil {
			return err
		}
		printer.Print([]map[string]string{
			{
				"host":          evt.NewMasterHost,
				"port":          evt.NewMasterPort,
				"previous_host": evt.OldMasterHost,
				"previous_port": evt.OldMasterPort,
			},
		}, []string{})
		break
	}
	return nil
}
