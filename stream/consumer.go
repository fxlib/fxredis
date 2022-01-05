package stream

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Delegate is provided to a consumer to handle an entry
type Delegate interface {
	HandleEntry(map[string]interface{}) error
}

// DelegateFunc implements the delegate
type DelegateFunc func(map[string]interface{}) error

// HandleEntry is called by the consumer to handle the entry
func (f DelegateFunc) HandleEntry(e map[string]interface{}) error {
	return f(e)
}

// Consumer pulls entries from a stream, identified as a consumer group.
type Consumer struct {
	cfg  ConsumerConf
	logs *zap.Logger
	rc   *redis.Client
	del  Delegate
}

// NewConsumer inits the consumer
func NewConsumer(
	lc fx.Lifecycle,
	cfg ConsumerConf,
	del Delegate,
	rc *redis.Client,
	logs *zap.Logger,
) (c *Consumer) {
	c = &Consumer{cfg: cfg, rc: rc, logs: logs, del: del}
	lc.Append(fx.Hook{OnStart: c.Start, OnStop: c.Stop})
	return
}

// Start the consumer by registering the group if it doesn't exist yet
func (c *Consumer) Start(ctx context.Context) error {
	return nil
}

// Stop the consumer group
func (c *Consumer) Stop(ctx context.Context) error {
	return nil
}
