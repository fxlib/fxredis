package stream

import (
	"os"
	"testing"

	"github.com/fxlib/fxredis"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestMultiConsumerEnvParsing(t *testing.T) {
	os.Setenv("C1_GROUP_NAME", "group1")
	defer os.Unsetenv("C1_GROUP_NAME")
	os.Setenv("C2_GROUP_NAME", "group2")
	defer os.Unsetenv("C2_GROUP_NAME")

	del1, del2 :=
		DelegateFunc(func(m map[string]interface{}) error { return nil }),
		DelegateFunc(func(m map[string]interface{}) error { return nil })
	var cs struct {
		fx.In
		C1 *Consumer `name:"c1"`
		C2 *Consumer `name:"c2"`
	}

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
		fx.Populate(&cs),
	).RequireStart().RequireStop()

	require.Equal(t, "group1", cs.C1.cfg.GroupName)
	require.Equal(t, "group2", cs.C2.cfg.GroupName)

	// @TODO test XADD resulting in entries being handled by delegates
	// @TODO test shutdown with clean-up to correctly reset after every test

}
