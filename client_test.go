package fxredis

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestClient(t *testing.T) {
	var rc *redis.Client
	fxtest.New(t,
		fx.Provide(New, zap.NewDevelopment, ParseEnv),
		fx.Supply(DefaultEnvPrefix),
		fx.Populate(&rc)).
		RequireStart().
		RequireStop()
}
