package cmd

import (
	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var sentinelWatchCmd = &cobra.Command{
	Use:   "watch [pattern (default *)]",
	Short: "Watch all events on the sentinel",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pattern := "*"
		if len(args) == 1 {
			pattern = args[0]
		}
		ExecuteSentinelWatch(&cfg, &redisCfg, prtr, pattern)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelWatchCmd)
}

func ExecuteSentinelWatch(
	config *config.RRTConfig,
	redisConfig *config.RedisConfig,
	printer *printer.Printer,
	pattern string,
) error {
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	pubsub := rdb.PSubscribe(ctx, pattern)
	defer pubsub.Close()
	// just print all the messages, without headers
	ch := pubsub.Channel()
	prtr.SkipHeaders = true
	prtr.Itemise = true
	for msg := range ch {
		prtr.Print([]map[string]string{
			{
				"ch":  msg.Channel,
				"msg": msg.Payload,
			},
		}, []string{})
	}
	return nil
}
