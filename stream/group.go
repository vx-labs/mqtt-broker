package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cenkalti/backoff"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	kv "github.com/vx-labs/mqtt-broker/kv/pb"
)

var (
	ErrShardLocked          = errors.New("shard locked")
	ErrShardAlreadyLocked   = errors.New("shard already locked")
	ErrShardAlreadyUnlocked = errors.New("shard already unlocked")
)

func newConsumerGroupConfig() consumerGroupConfig {
	return consumerGroupConfig{
		ConsumerHeartbeats: map[string]time.Time{},
	}

}

type consumerGroupConfig struct {
	ConsumerHeartbeats map[string]time.Time `json:"consumer_heartbeats"`
}

func updateKeyValue(ctx context.Context, client *kv.Client, key []byte, f func([]byte, []byte) ([]byte, error)) error {
	return backoff.Retry(func() error {
		value, md, err := client.GetWithMetadata(ctx, key)
		if err != nil {
			return backoff.Permanent(err)
		}
		newValue, err := f(key, value)
		if err != nil {
			return backoff.Permanent(err)
		}
		if newValue == nil {
			return nil
		}
		err = client.SetWithVersion(ctx, key, newValue, md.Version, kv.WithTimeToLive(60*time.Second))
		if err != nil {
			if grpcCode := status.Code(err); grpcCode == codes.FailedPrecondition {
				return err
			}
			return backoff.Permanent(err)
		}
		return nil
	}, backoff.NewExponentialBackOff())
}

func getAssignedShard(shards, consumers []string, self string) []string {
	sort.SliceStable(consumers, func(a, b int) bool {
		return strings.Compare(consumers[a], consumers[b]) == -1
	})
	sort.SliceStable(shards, func(a, b int) bool {
		return strings.Compare(shards[a], shards[b]) == -1
	})
	selfIdx := sort.SearchStrings(consumers, self)

	out := []string{}
	for idx := range shards {
		if (idx)%(len(consumers)) == selfIdx {
			out = append(out, shards[idx])
		}
	}
	return out
}
func listConsumers(ctx context.Context, client *kv.Client, streamID, groupID string) ([]string, error) {
	key := []byte(fmt.Sprintf("stream/%s/%s/config", streamID, groupID))
	value, err := client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	config := newConsumerGroupConfig()
	err = json.Unmarshal(value, &config)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(config.ConsumerHeartbeats))
	for consumer := range config.ConsumerHeartbeats {
		out = append(out, consumer)
	}
	return out, nil
}
func heartbeatConsumer(ctx context.Context, client *kv.Client, streamID, groupID, consumerID string) error {
	key := []byte(fmt.Sprintf("stream/%s/%s/config", streamID, groupID))
	now := time.Now()
	return updateKeyValue(ctx, client, key, func(key, value []byte) ([]byte, error) {
		config := newConsumerGroupConfig()
		if len(value) != 0 {
			err := json.Unmarshal(value, &config)
			if err != nil {
				return nil, err
			}
		}
		config.ConsumerHeartbeats[consumerID] = now
		updatedConfig, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		return updatedConfig, nil
	})
}
func purgeDeadConsumers(ctx context.Context, client *kv.Client, streamID, groupID string) error {
	key := []byte(fmt.Sprintf("stream/%s/%s/config", streamID, groupID))
	deadline := 10 * time.Second
	now := time.Now()
	return updateKeyValue(ctx, client, key, func(key, value []byte) ([]byte, error) {
		config := newConsumerGroupConfig()
		if len(value) != 0 {
			err := json.Unmarshal(value, &config)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, nil
		}
		deadConsumers := []string{}
		for consumer, heartbeat := range config.ConsumerHeartbeats {
			if now.Sub(heartbeat) >= deadline {
				deadConsumers = append(deadConsumers, consumer)
			}
		}
		for _, deadConsumer := range deadConsumers {
			delete(config.ConsumerHeartbeats, deadConsumer)
		}
		updatedConfig, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		return updatedConfig, nil
	})
}

func registerConsumer(ctx context.Context, client *kv.Client, streamID, groupID, consumerID string) error {
	key := []byte(fmt.Sprintf("stream/%s/%s/config", streamID, groupID))
	return updateKeyValue(ctx, client, key, func(key, value []byte) ([]byte, error) {
		config := newConsumerGroupConfig()
		if len(value) != 0 {
			err := json.Unmarshal(value, &config)
			if err != nil {
				return nil, err
			}
		}
		config.ConsumerHeartbeats[consumerID] = time.Now()
		updatedConfig, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		return updatedConfig, nil
	})
}

func unregisterConsumer(ctx context.Context, client *kv.Client, streamID, groupID, consumerID string) error {
	key := []byte(fmt.Sprintf("stream/%s/%s/config", streamID, groupID))
	return updateKeyValue(ctx, client, key, func(key, value []byte) ([]byte, error) {
		config := newConsumerGroupConfig()
		if len(value) != 0 {
			err := json.Unmarshal(value, &config)
			if err != nil {
				return nil, err
			}
		}
		delete(config.ConsumerHeartbeats, consumerID)
		updatedConfig, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		return updatedConfig, nil
	})
}

func unlockShard(ctx context.Context, client *kv.Client, streamID, groupID, shardID, consumerID string) error {
	key := []byte(fmt.Sprintf("stream/%s/%s/%s/lock", streamID, shardID, groupID))
	value, md, err := client.GetWithMetadata(ctx, key)
	if err != nil {
		return err
	}
	if len(value) == 0 {
		return ErrShardAlreadyUnlocked
	}
	if string(value) != consumerID {
		return ErrShardLocked
	}

	return client.DeleteWithVersion(ctx, key, md.Version)

}

func lockShard(ctx context.Context, client *kv.Client, streamID, groupID, shardID, consumerID string, duration time.Duration) (uint64, error) {
	key := []byte(fmt.Sprintf("stream/%s/%s/%s/lock", streamID, shardID, groupID))
	value, md, err := client.GetWithMetadata(ctx, key)
	if err != nil {
		return 0, err
	}
	if len(value) != 0 {
		return 0, ErrShardAlreadyLocked
	}
	return md.Version + 1, client.SetWithVersion(ctx, key, []byte(consumerID), md.Version, kv.WithTimeToLive(duration))
}
func renewShardlock(ctx context.Context, client *kv.Client, streamID, groupID, shardID, consumerID string, duration time.Duration, version uint64) error {
	key := []byte(fmt.Sprintf("stream/%s/%s/%s/lock", streamID, shardID, groupID))
	return client.SetWithVersion(ctx, key, []byte(consumerID), version, kv.WithTimeToLive(duration))
}
