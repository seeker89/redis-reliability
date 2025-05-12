package redisClient

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

type RedisInstance struct {
	Host   string
	Port   string
	Master string
}

type RedisSwitchMasterEvent struct {
	Master        string
	NewMasterHost string
	NewMasterPort string
	OldMasterHost string
	OldMasterPort string
}

func MakeRedisClient(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opts), nil
}

func GetMasterFromSentinel(rdb *redis.Client, ctx context.Context, master string) (*RedisInstance, error) {
	cmd := rdb.Do(ctx, "SENTINEL", "get-master-addr-by-name", master)
	if err := rdb.Process(ctx, cmd); err != nil {
		return nil, err
	}
	res, err := cmd.StringSlice()
	if err != nil {
		return nil, err
	}
	return &RedisInstance{
		Host:   res[0],
		Port:   res[1],
		Master: master,
	}, nil
}

func ParseSwitchMasterMessage(msg string) (*RedisSwitchMasterEvent, error) {
	parts := strings.Split(msg, " ")
	if len(parts) != 5 {
		return nil, fmt.Errorf("expected formatted redis +switch-master, got %s", msg)
	}
	return &RedisSwitchMasterEvent{
		Master:        parts[0],
		OldMasterHost: parts[1],
		OldMasterPort: parts[2],
		NewMasterHost: parts[3],
		NewMasterPort: parts[4],
	}, nil
}

func WaitForNewMaster(
	ctx context.Context,
	rdbs *redis.Client,
	done chan error,
	pq chan map[string]string,
	oldMaster *RedisInstance,
) {
	// listen in on new sentinel events
	spubsub := rdbs.PSubscribe(ctx, "+*")
	defer spubsub.Close()
	for msg := range spubsub.Channel() {
		switch msg.Channel {
		case "+switch-master":
			evt, err := ParseSwitchMasterMessage(msg.Payload)
			if err != nil {
				pq <- map[string]string{
					"event": "bad message",
					"msg":   err.Error(),
				}
				done <- err
				continue
			}
			// ignore if the message for different master
			if evt.Master != oldMaster.Master {
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
			newMaster, err := GetMasterFromSentinel(rdbs, ctx, oldMaster.Master)
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
}
