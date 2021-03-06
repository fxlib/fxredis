package fxredis

import (
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestClient(t *testing.T) {
	var rc *redis.Client
	fxtest.New(t,
		fx.Supply(env.Options{}),
		fx.Provide(New, zap.NewDevelopment, ParseEnv),
		fx.Populate(&rc)).
		RequireStart().
		RequireStop()
}

type testDep1 struct{ c1, c2 *redis.Client }

func newMultiConnTest(c1, c2 *redis.Client) *testDep1 {
	return &testDep1{c1, c2}
}

func newInvokedTest(c1, c2 *redis.Client) {}

func TestMultipleClients(t *testing.T) {
	var dep1 *testDep1
	var clients struct {
		fx.In
		C1 *redis.Client `name:"c1"`
		C2 *redis.Client `name:"c2"`
	}

	defer fxtest.New(t,
		fx.Populate(&dep1, &clients),
		fx.Supply(
			fx.Annotate(env.Options{Prefix: "C1_"}, fx.ResultTags(`name:"env_c1"`)),
			fx.Annotate(env.Options{Prefix: "C2_"}, fx.ResultTags(`name:"env_c2"`)),
		),
		fx.Provide(
			fx.Annotate(ParseEnv, fx.ParamTags(`name:"env_c1"`), fx.ResultTags(`name:"c1"`)),
			fx.Annotate(ParseEnv, fx.ParamTags(`name:"env_c2"`), fx.ResultTags(`name:"c2"`)),
			fx.Annotate(New, fx.ParamTags(``, `name:"c1"`), fx.ResultTags(`name:"c1"`)),
			fx.Annotate(New, fx.ParamTags(``, `name:"c2"`), fx.ResultTags(`name:"c2"`)),
			fx.Annotate(newMultiConnTest, fx.ParamTags(`name:"c1"`, `name:"c2"`)),
		),
		fx.Invoke(
			fx.Annotate(newInvokedTest, fx.ParamTags(`name:"c1"`, `name:"c2"`))),
	).RequireStart().RequireStop()

	require.NotNil(t, dep1)
	require.NotNil(t, clients.C1)
	require.NotNil(t, clients.C2)
}
