package fxredis

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-redis/redis/v8"
)

// Options is a type alias the can be parsed as an evironment option. Check out the test file on how
// this can be used for parsing environment variables.
type Options redis.Options

// UnmarshalText allows the options to be unmarshalled
func (opts *Options) UnmarshalText(text []byte) (err error) {
	ropts, err := redis.ParseURL(string(text))
	if err == nil {
		*opts = Options(*ropts)
	}
	return
}

// Conf can be filled by parsing environment variables. It is common that this is prefixed and included
// as a field in a larger configuration.
type Conf struct {
	// URL must be set in the environment as a basic connection string
	URL *Options `env:"URL" envDefault:"redis://localhost:6379"`
	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int `env:"MAX_RETRIES"`
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration `env:"MIN_RETRY_BACKOFF"`
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration `env:"MAX_RETRY_BACKOFF"`
	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `env:"DIAL_TIMEOUT"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration `env:"READ_TIMEOUT"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT"`
}

// ConfToOpts can convert a parsed Conf to options. It is exposed in case it is necessary to parse the
// Conf struct manually.
func ConfToOpts(cfg Conf) *redis.Options {
	cfg.URL.MaxRetries = cfg.MaxRetries
	cfg.URL.MinRetryBackoff = cfg.MinRetryBackoff
	cfg.URL.MaxRetryBackoff = cfg.MaxRetryBackoff
	cfg.URL.DialTimeout = cfg.DialTimeout
	cfg.URL.ReadTimeout = cfg.ReadTimeout
	cfg.URL.WriteTimeout = cfg.WriteTimeout
	ropts := redis.Options(*cfg.URL)
	return &ropts
}

// ParseEnv will parse environment variables into Redis options. The parsing can be custimized by configuring
// a prefix.
func ParseEnv(prefix EnvPrefix) (opts *redis.Options, err error) {
	var cfg Conf
	if err = env.Parse(&cfg, env.Options{Prefix: string(prefix)}); err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}
	return ConfToOpts(cfg), err
}

// EnvPrefix can be used to declaratively prefix the parsing of Redis env configuration
type EnvPrefix string

// DefaultEnvPrefix is a common env prefix
var DefaultEnvPrefix EnvPrefix = "REDIS_"
