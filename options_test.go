package fxredis

import (
	"os"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/require"
)

func TestOptionUnmarshal(t *testing.T) {
	var cfg struct {
		Opts *Options `env:"OPTIONS" envDefault:"redis://foo:5000/100"`
	}

	require.NoError(t, env.Parse(&cfg))
	require.Equal(t, "foo:5000", cfg.Opts.Addr)
	require.Equal(t, 100, cfg.Opts.DB)
}

func TestOptionParsing(t *testing.T) {
	os.Setenv("REDIS_URL", "redis://bar:6000")
	os.Setenv("REDIS_MAX_RETRIES", "5")
	defer os.Unsetenv("REDIS_URL")
	defer os.Unsetenv("REDIS_MAX_RETRIES")

	opts, err := ParseEnv(env.Options{Prefix: "REDIS_"})
	require.NoError(t, err)
	require.Equal(t, "bar:6000", opts.Addr)
}

func TestParsingError(t *testing.T) {
	os.Setenv("REDIS_URL", "redis:/bar:6000")
	defer os.Unsetenv("REDIS_URL")

	_, err := ParseEnv(env.Options{Prefix: "REDIS_"})
	require.Error(t, err)
}
