package stream

import (
	"context"
	"os"
	"testing"

	"github.com/fxlib/fxredis"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestMultiConsumerEnvParsing(t *testing.T) {
	os.Setenv("C1_GROUP_NAME", "group1")
	defer os.Unsetenv("C1_GROUP_NAME")
	os.Setenv("C1_STREAM_NAMES", "test_stream1")
	defer os.Unsetenv("C1_STREAM_NAMES")
	os.Setenv("C2_GROUP_NAME", "group2")
	defer os.Unsetenv("C2_GROUP_NAME")
	os.Setenv("C2_STREAM_NAMES", "test_stream1")
	defer os.Unsetenv("C2_STREAM_NAMES")

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

	// Use fx to setup two consumers with separate delegates
	defer fxtest.New(t,
		fx.Provide(zap.NewDevelopment, fxredis.New, fxredis.ParseEnv),
		fx.Supply(fx.Annotate(del1, fx.ResultTags(`name:"c1"`), fx.As(new(Delegate)))),
		fx.Supply(fx.Annotate(del2, fx.ResultTags(`name:"c2"`), fx.As(new(Delegate)))),
		fx.Supply(fx.Annotate([]string{"C1_"}, fx.ResultTags(`name:"c1_env_prefix"`))),
		fx.Supply(fx.Annotate([]string{"C2_"}, fx.ResultTags(`name:"c2_env_prefix"`))),
		fx.Provide(
			fx.Annotate(ParseConsumerEnv, fx.ResultTags(`name:"c1"`), fx.ParamTags(`name:"c1_env_prefix"`)),
			fx.Annotate(NewConsumer, fx.ParamTags(``, `name:"c1"`, `name:"c1"`), fx.ResultTags(`name:"c1"`)),
		),
		fx.Provide(
			fx.Annotate(ParseConsumerEnv, fx.ResultTags(`name:"c2"`), fx.ParamTags(`name:"c2_env_prefix"`)),
			fx.Annotate(NewConsumer, fx.ParamTags(``, `name:"c2"`, `name:"c2"`), fx.ResultTags(`name:"c2"`)),
		),
		fx.Populate(&cs, &rc),
	).RequireStart().RequireStop()

	require.Equal(t, "group1", cs.C1.cfg.GroupName)
	require.Equal(t, "group2", cs.C2.cfg.GroupName)

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
	os.Setenv("GROUP_NAME", "") // this will cause the consumer to always fail reading
	defer os.Unsetenv("GROUP_NAME")

	zc, obs := observer.New(zap.DebugLevel)

	del := DelegateFunc(func(stream, mid string, values map[string]interface{}) error { return nil })
	var csm *Consumer
	fxtest.New(t,
		fx.Provide(fxredis.New, fxredis.ParseEnv, ParseConsumerEnv, NewConsumer, zap.New),
		fx.Supply(fx.Annotate(zc, fx.As(new(zapcore.Core)))),
		fx.Supply(fx.Annotate(del, fx.As(new(Delegate)))),
		fx.Populate(&csm),
	).RequireStart().RequireStop()

	require.Equal(t, 1, obs.FilterMessage(`shutting down waiting for backoff`).Len())
}
