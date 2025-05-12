package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/seeker89/redis-resiliency-toolkit/pkg/config"
	"github.com/seeker89/redis-resiliency-toolkit/pkg/k8s"
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
	printer.SkipHeaders = true
	printer.Itemise = true
	printOne := func(data map[string]string) {
		data["time"] = time.Now().String()
		printer.Print([]map[string]string{data}, []string{"time", "event", "ch", "msg"})
	}
	pq := make(chan map[string]string, 10)
	go func() {
		for {
			printOne(<-pq)
		}
	}()
	rdbs, err := redisClient.MakeRedisClient(redisConfig.SentinelURL)
	if err != nil {
		return err
	}
	done := make(chan error)
	// The plan here is:
	// 1. read the master from sentinel
	// 2. query INFO from the master to see that it matches what sentinel gave us
	//    by default, use the host:port from the sentinel
	//    alternatively, use specified ingress/proxy
	// 3. set up a sentinel event watcher & specified timeout
	// 4. kill the pod containing current master
	//    continue killing if the master switchover hasn't happened
	// 6. wait for either 1) timeout, or 2) +switch-master event
	// 7. read the master from sentinel again
	// 8. query INFO from the master again
	oldMaster, err := redisClient.GetMasterFromSentinel(ctx, rdbs, redisConfig.SentinelMaster)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	pq <- map[string]string{
		"event": "initial master",
		"msg":   fmt.Sprintf("%s:%s", oldMaster.Host, oldMaster.Port),
	}
	go redisClient.WaitForNewMaster(
		ctx,
		rdbs,
		done,
		pq,
		oldMaster,
		cfg.Verbose,
	)
	// keep the pod dead until we succeed
	k8sc, err := k8s.GetClient(config.Kubeconfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	n, err := k8s.GuessPodNameFromHost(oldMaster.Host)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	pq <- map[string]string{
		"event": "pod name",
		"msg":   n,
	}
	go k8s.KeepPodDead(ctx, k8sc, n, cfg.Namespace, done, pq)
	// add a max timeout for all of this
	go func() {
		timeout := 60 * time.Second
		time.Sleep(timeout)
		pq <- map[string]string{
			"event":    "timeout",
			"duration": timeout.String(),
		}
		done <- fmt.Errorf("timeout after %s", timeout)
	}()
	return <-done
}
