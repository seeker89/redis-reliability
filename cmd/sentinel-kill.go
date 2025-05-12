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
	oldMaster, err := redisClient.GetMasterFromSentinel(rdbs, ctx, redisConfig.SentinelMaster)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	pq <- map[string]string{
		"event": "initial master",
		"msg":   fmt.Sprintf("%s:%s", oldMaster.Host, oldMaster.Port),
	}
	// listen in on new sentinel events
	spubsub := rdbs.PSubscribe(ctx, "+*")
	defer spubsub.Close()
	go func() {
		for msg := range spubsub.Channel() {
			switch msg.Channel {
			case "+switch-master":
				evt, err := redisClient.ParseSwitchMasterMessage(msg.Payload)
				if err != nil {
					pq <- map[string]string{
						"event": "bad message",
						"msg":   err.Error(),
					}
					done <- err
					continue
				}
				// ignore if the message for different master
				if evt.Master != redisConfig.SentinelMaster {
					pq <- map[string]string{
						"event":  "different master",
						"master": evt.Master,
					}
					continue
				}
				// final check
				if evt.OldMasterHost != oldMaster.Host || evt.OldMasterPort != oldMaster.Port {
					done <- fmt.Errorf("previous master doesn't match; got %v, wanted %v", evt, oldMaster)
					break
				}
				newMaster, err := redisClient.GetMasterFromSentinel(rdbs, ctx, redisConfig.SentinelMaster)
				if err != nil {
					done <- err
					break
				}
				if evt.NewMasterHost != newMaster.Host || evt.NewMasterPort != newMaster.Port {
					done <- fmt.Errorf("new master doesn't match; got %v, wanted %v", newMaster, evt)
					break
				}
				pq <- map[string]string{
					"event":      "done",
					"result":     "OK",
					"new_master": fmt.Sprintf("%s:%s", evt.NewMasterHost, evt.NewMasterPort),
				}
				done <- nil
			}
			pq <- map[string]string{
				"event": "sentinel message",
				"ch":    msg.Channel,
				"msg":   msg.Payload,
			}
		}
	}()
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
