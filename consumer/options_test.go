package consumer

import (
	"os"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestParseEnvOptions(t *testing.T) {
	var o Options
	opts, err := ParseEnv(env.Options{})
	require.NoError(t, err)
	require.Len(t, opts, 1)
	opts[0](&o)

	require.Equal(t, time.Second, o.BlockTime)
}

func TestDefaultOptions(t *testing.T) {
	os.Setenv("BLOCK_TIME", "2s") // this should be ignored
	defer os.Unsetenv("BLOCK_TIME")

	logs, _ := zap.NewDevelopment()
	lc := fxtest.NewLifecycle(t)

	t.Run("without options should have default from env struct tags", func(t *testing.T) {
		c := New(lc, logs, nil, nil, "foo")
		require.Equal(t, time.Second, c.opts.BlockTime)
	})

	t.Run("but options should overwrite this", func(t *testing.T) {
		c := New(lc, logs, nil, nil, "foo", BlockTime(time.Second*3))
		require.Equal(t, time.Second*3, c.opts.BlockTime)
	})
}

func TestOptions(t *testing.T) {
	var opts Options
	for _, o := range []Option{
		StreamNames("foo", "bar"),
		StreamStart("100"),
		ReadPerBlock(4),
		HandleTime(time.Second * 40),
	} {
		o(&opts)
	}

	require.Equal(t, Options{
		StreamNames:  []string{"foo", "bar"},
		StreamStart:  "100",
		ReadPerBlock: 4,
		HandleTime:   time.Second * 40,
	}, opts)
}
