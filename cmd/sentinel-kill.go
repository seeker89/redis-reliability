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
		return ExecuteSentinelKill(&cfg, prtr)
	},
}

func init() {
	sentinelCmd.AddCommand(sentinelKillCmd)
}

func ExecuteSentinelKill(
	config *config.RRTConfig,
	printer *printer.Printer,
) error {
	// we'll be emitting events one by one
	printer.SkipHeaders = true
	printer.Itemise = true
	// we'll write here the messages to print
	pq := make(chan map[string]string, 10)
	pqdone := make(chan bool)
	go func() {
		for {
			data := <-pq
			if data["debug"] != "" && !config.Verbose {
				continue
			}
			delete(data, "debug")
			data["time"] = time.Now().String()
			printer.Print([]map[string]string{data}, []string{"time", "event", "msg"})
			if data["done"] == "true" {
				pqdone <- true
			}
		}
	}()

	// The plan here is:
	// 1. read the master from sentinel
	// 2. query INFO from the master to see that it matches what sentinel gave us
	//    by default, use the host:port from the sentinel
	//    alternatively, use specified ingress/proxy
	// 3. set up a sentinel event watcher
	// 4. kill the pod containing current master
	//    continue killing if the master switchover hasn't happened
	// 5. setup the maximum timeout
	// 7. read the master from sentinel again
	// 8. query INFO from the master again

	rdbs, err := redisClient.MakeRedisClient(config.SentinelURL)
	if err != nil {
		return err
	}
	done := make(chan error)

	// 1. Read the old master from the sentinel
	oldMaster, err := redisClient.GetMasterFromSentinel(ctx, rdbs, config.SentinelMaster)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	pq <- map[string]string{
		"event": "initial master",
		"msg":   fmt.Sprintf("%s:%s", oldMaster.Host, oldMaster.Port),
	}

	// 3. Listen to sentinel events, and finish early when possible
	go redisClient.WaitForNewMaster(
		ctx,
		rdbs,
		done,
		pq,
		oldMaster,
	)

	// 4. Keep killing the pods without grace period
	k8sc, err := k8s.GetClient(config.Kubeconfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	ns := k8s.DeriveNamespace(cfg.Namespace)
	n, err := k8s.GuessPodNameFromHost(oldMaster.Host)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	go k8s.KeepPodDead(
		ctx,
		k8sc,
		n,
		ns,
		int64(config.Grace.Seconds()),
		done,
		pq,
	)

	// 5. Setup the max time this all should take
	go func(timeout time.Duration) {
		time.Sleep(timeout)
		pq <- map[string]string{
			"event":    "timeout",
			"duration": timeout.String(),
		}
		done <- fmt.Errorf("timeout after %s", timeout)
	}(config.Timeout)

	// wait for the race to end
	result := <-done

	// 7. Read the master again from the sentinel
	newMaster, err := redisClient.GetMasterFromSentinel(ctx, rdbs, config.SentinelMaster)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	pq <- map[string]string{
		"done":  "true",
		"event": "final master",
		"msg":   fmt.Sprintf("%s:%s", newMaster.Host, newMaster.Port),
	}

	// wait up for any in-transit messages
	<-pqdone

	return result
}
