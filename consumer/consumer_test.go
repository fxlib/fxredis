package consumer

import (
	"context"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/fxlib/fxredis"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// provideConsumer provides a consumer under the provided name 's', using delegate 'del' and options 'opts'
func provideConsumer(s string, del Delegate, opts ...Option) fx.Option {
	return fx.Options(
		fx.Supply(fx.Annotate(opts, fx.ResultTags(`name:"`+s+`"`))),
		fx.Supply(fx.Annotate(s, fx.ResultTags(`name:"`+s+`_group_name"`))),
		fx.Supply(fx.Annotate(del, fx.ResultTags(`name:"`+s+`"`), fx.As(new(Delegate)))),
		fx.Provide(
			fx.Annotate(New,
				fx.ParamTags(``, ``, ``, `name:"`+s+`"`, `name:"`+s+`_group_name"`, `name:"`+s+`"`),
				fx.ResultTags(`name:"`+s+`"`)),
		),
	)
}

func TestMultipleConsumerDI(t *testing.T) {
	delc := make(chan string, 255)
	del1, del2 :=
		DelegateFunc(func(stream, mid string, values map[string]interface{}) error { delc <- mid; return nil }),
		DelegateFunc(func(stream, mid string, values map[string]interface{}) error { delc <- mid; return nil })
	var rc *redis.Client
	var cs struct {
		fx.In
		C1 *Consumer `name:"c1"`
		C2 *Consumer `name:"c2"`
	}

	defer fxtest.New(t,
		fx.Supply(env.Options{}),
		fx.Populate(&cs, &rc),
		fx.Provide(zap.NewDevelopment, fxredis.New, fxredis.ParseEnv),
		provideConsumer("c1", del1, StreamNames("test_stream1")),
		provideConsumer("c2", del2, StreamNames("test_stream1")),
	).RequireStart().RequireStop()

	require.Equal(t, "c1", cs.C1.group)
	require.Equal(t, "c2", cs.C2.group)

	ctx := context.Background()
	t.Run("xadd to stream should result in both delegates being called", func(t *testing.T) {
		args := &redis.XAddArgs{ID: "*", Stream: "test_stream1", Values: map[string]interface{}{
			"id":          1,
			"rel_id":      100,
			"schema_name": "public",
			"table_name":  "sea_services",
			"tstamp_tx":   "2022-01-04T11:37:57+01:00",
			"op":          "U",
			"old":         `{"foo":1}`,
			"new":         `{"bar":2}`,
		}}

		require.NoError(t, rc.XAdd(ctx, args).Err())
		require.Contains(t, <-delc, "-")
		require.Contains(t, <-delc, "-")
	})
}

func TestFailingConsumer(t *testing.T) {
	zc, obs := observer.New(zap.DebugLevel)

	del := DelegateFunc(func(stream, mid string, values map[string]interface{}) error { return nil })
	var csm *Consumer
	fxtest.New(t,
		fx.Supply(env.Options{}),
		fx.Supply([]Option{}),
		fx.Supply(fx.Annotate("", fx.ResultTags(`name:"group_name"`))),
		fx.Provide(fxredis.New, fxredis.ParseEnv, zap.New),
		fx.Provide(fx.Annotate(New, fx.ParamTags(``, ``, ``, ``, `name:"group_name"`))),
		fx.Supply(fx.Annotate(zc, fx.As(new(zapcore.Core)))),
		fx.Supply(fx.Annotate(del, fx.As(new(Delegate)))),
		fx.Populate(&csm),
	).RequireStart().RequireStop()

	require.Equal(t, 1, obs.FilterMessage(`shutting down waiting for backoff`).Len())
}
