package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultQueueKey = "statesight:jobs:default"

type RedisQueue struct {
	client      *redis.Client
	queueKey    string
	pollTimeout time.Duration
}

func NewRedisQueue(redisURL string, pollTimeout time.Duration) (*RedisQueue, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(options)

	return &RedisQueue{
		client:      client,
		queueKey:    defaultQueueKey,
		pollTimeout: pollTimeout,
	}, nil
}

func (q *RedisQueue) Enqueue(ctx context.Context, msg Message) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	body, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal job message: %w", err)
	}
	if err := q.client.LPush(ctx, q.queueKey, body).Err(); err != nil {
		return fmt.Errorf("redis lpush: %w", err)
	}
	return nil
}

func (q *RedisQueue) Consume(ctx context.Context, handler Handler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, err := q.client.BRPop(ctx, q.pollTimeout, q.queueKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			// During shutdown we can get context cancellation from BRPOP.
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("redis brpop: %w", err)
		}
		if len(result) != 2 {
			continue
		}

		msg, err := UnmarshalMessage([]byte(result[1]))
		if err != nil {
			return fmt.Errorf("decode queue message: %w", err)
		}
		if err := handler(ctx, msg); err != nil {
			return err
		}
	}
}

func (q *RedisQueue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}
