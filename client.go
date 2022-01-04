package fxredis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
)

// New inits a normal Redis client but it will automatically ping the Redis connection when
// started. It will also close the connection when the fx application stops.
func New(lc fx.Lifecycle, opts *redis.Options) (rc *redis.Client) {
	rc = redis.NewClient(opts)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return rc.Ping(ctx).Err() },
		OnStop:  func(ctx context.Context) error { return rc.Close() },
	})
	return
}
