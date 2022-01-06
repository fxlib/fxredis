package consumer

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Delegate is provided to a consumer to handle an entry
type Delegate interface {
	HandleMessage(ctx context.Context, stream, mid string, values map[string]interface{}) error
}

// DelegateFunc implements the delegate
type DelegateFunc func(ctx context.Context, stream, mid string, values map[string]interface{}) error

// HandleMessage is called by the consumer to handle the entry
func (f DelegateFunc) HandleMessage(ctx context.Context, stream, mid string, values map[string]interface{}) error {
	return f(ctx, stream, mid, values)
}

// Consumer pulls entries from a stream, identified as a consumer group.
type Consumer struct {
	logs  *zap.Logger
	rc    *redis.Client
	del   Delegate
	group string

	opts  Options
	close uint32
	done  chan struct{}
	endbo chan struct{}
}

// New inits the consumer
func New(
	lc fx.Lifecycle,
	logs *zap.Logger,
	rc *redis.Client,
	del Delegate,
	group string,
	opts ...Option,
) (c *Consumer) {
	c = &Consumer{
		rc:    rc,
		logs:  logs.With(zap.String("group_name", group)),
		del:   del,
		group: group,
		done:  make(chan struct{}, 1), endbo: make(chan struct{}, 1),
	}

	// defaults as defined through the envDefault tag, then overwrite with any options
	env.Parse(&c.opts, env.Options{Environment: map[string]string{}})
	for _, o := range opts {
		o(&c.opts)
	}

	lc.Append(fx.Hook{OnStart: c.Start, OnStop: c.Stop})
	return
}

// Start the consumer by registering the group if it doesn't exist yet
func (c *Consumer) Start(ctx context.Context) error {
	for _, sname := range c.opts.StreamNames {
		c.logs.Info("ensuring consumer group exists",
			zap.String("stream_name", sname))

		// setup the groups and make streams if necessary
		if err := c.rc.XGroupCreateMkStream(
			ctx, sname, c.group, c.opts.StreamStart,
		).Err(); err != nil && strings.Contains(err.Error(), "BUSYGROUP") {
			c.logs.Info("consumer group already exists, do nothing",
				zap.String("stream_name", sname))
		} else if err != nil {
			return fmt.Errorf("failed to ensure consume group exists: %w", err)
		}
	}

	go c.handleMessages()
	return nil
}

// handleNextMessage will block and wait until some message is available
func (c *Consumer) handleNextMessage() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.opts.HandleTime)
	defer cancel()

	streamarg := make([]string, len(c.opts.StreamNames)*2)
	for i, sn := range c.opts.StreamNames {
		streamarg[i] = sn
		streamarg[i+1] = ">"
	}

	args := &redis.XReadGroupArgs{
		Group:   c.group,
		Streams: streamarg,
		Count:   c.opts.ReadPerBlock,
		Block:   c.opts.BlockTime,
		NoAck:   false,
	}

	res, err := c.rc.XReadGroup(ctx, args).Result()
	c.logs.Debug("called bocking XReadGroup",
		zap.Error(err), zap.Int("res_len", len(res)), zap.Any("args", args))

	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to call XReadGroup: %w", err)
	}

	// pass messages from each stream over to the delegate. If any of them fail
	// the message is not ACK
	for _, stream := range res {
		for _, msg := range stream.Messages {
			c.logs.Debug("delegating message",
				zap.String("stream_name", stream.Stream),
				zap.String("message_id", msg.ID),
				zap.Any("values", msg.Values))

			if err := c.del.HandleMessage(ctx, stream.Stream, msg.ID, msg.Values); err != nil {
				c.logs.Error("delegate failed to handle message, skipping ACK",
					zap.Error(err),
					zap.String("stream_name", stream.Stream),
					zap.String("message_id", msg.ID),
					zap.Any("values", msg.Values))
				continue
			}

			if err := c.rc.XAck(ctx, stream.Stream, c.group, msg.ID).Err(); err != nil {
				return fmt.Errorf("failed to XACk message '%s': %w", msg.ID, err)
			}
		}
	}

	return nil
}

// handleMessages runs the message reading loop
func (c *Consumer) handleMessages() {
	defer close(c.done)
	bo := backoff.NewExponentialBackOff()

	for {
		err := c.handleNextMessage()
		if err == nil {
			if atomic.LoadUint32(&c.close) > 0 {
				return // shutting down the consumer
			}

			bo.Reset()
			continue //ok, continue without delay
		}

		// backoff if we retry too quickly
		wait := bo.NextBackOff()
		c.logs.Error("failed to handle next message, backoff",
			zap.Error(err),
			zap.Duration("backoff", wait))

		select {
		case <-time.After(wait):
		case <-c.endbo:
			c.logs.Debug("shutting down waiting for backoff")
			return //closing while waiting for backoff
		}
	}
}

// Stop the consumer group
func (c *Consumer) Stop(ctx context.Context) error {
	close(c.endbo)                  // signal any backoff to stop
	atomic.StoreUint32(&c.close, 1) // signal not to block another time

	c.logs.Info("waiting for consumer's message handling")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return nil
	}
}
