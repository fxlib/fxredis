package fxredis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger(t *testing.T) {
	var rl *Logger
	zc, obs := observer.New(zap.DebugLevel)

	fxtest.New(t,
		fx.Provide(NewLogger, zap.New, func() zapcore.Core { return zc }),
		fx.Populate(&rl))

	rl.Printf(context.Background(), "foo")
	require.Equal(t, 1, obs.FilterMessage("foo").Len())
}
