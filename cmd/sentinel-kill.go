package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/printer"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/redisClient"
	"github.com/spf13/cobra"
)

var sentinelKillCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill the master to trigger failover",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteSentinelKill(&cfg, &redisCfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelKillCmd)
}

func ExecuteSentinelKill(
	config *config.RRTConfig,
	redisConfig *config.RedisSentinelConfig,
	printer *printer.Printer,
) error {
	// we'll be emitting events one by one
	printer.Itemise = true
	printOne := func(data map[string]string) {
		data["time"] = time.Now().String()
		printer.Print([]map[string]string{data}, []string{"time", "event", "msg"})
	}
	rdb, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	} // The plan here is:
	// 1. read the master from sentinel
	// 2. query INFO from the master to see that it matches what sentinel gave us
	//    by default, use the host:port from the sentinel
	//    alternatively, use specified ingress/proxy
	// 3. set up the event watcher & specified timeout
	// 4. kill the pod containing current master
	//    continue killing if the master switchover hasn't happened
	// 6. wait for either 1) timeout, or 2) +switch-master event
	// 7. read the master from sentinel again
	// 8. query INFO from the master again
	master, err := redisClient.GetMasterFromSentinel(rdb, ctx, redisConfig.SentinelMaster)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	printOne(map[string]string{
		"event": "initial master",
		"msg":   fmt.Sprintf("%s:%s", master.Host, master.Port),
	})
	return nil
}
