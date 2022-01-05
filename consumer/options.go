package consumer

import (
	"time"

	"github.com/caarlos0/env/v6"
)

// ParseEnv will parse environment variables into options
func ParseEnv(eo env.Options) (opts []Option, err error) {
	var o Options
	opts = append(opts, FromOptions(&o))
	return opts, env.Parse(&o, eo)
}

// Options for the consumer with env keys configured
type Options struct {
	StreamNames  []string      `env:"STREAM_NAMES"`
	StreamStart  int           `env:"STREAM_START"`
	BlockTime    time.Duration `env:"BLOCK_TIME" envDefault:"1s"`
	ReadPerBlock int64         `env:"READ_PER_BLOCK" envDefault:"2"`
	HandleTime   time.Duration `env:"HANDLE_TIME" envDefault:"10s"`
}

// Option configures the consumer
type Option func(c *Options)

// StreamNames configures names of the streams the consumer reads from
func StreamNames(v ...string) Option {
	return func(c *Options) { c.StreamNames = v }
}

// StreamStart will determine at which index the consumer starts
func StreamStart(v int) Option {
	return func(c *Options) { c.StreamStart = v }
}

// BlockTime configures how long each read will block to wait for messages
func BlockTime(v time.Duration) Option {
	return func(c *Options) { c.BlockTime = v }
}

// ReadPerBlock configures how many messages to read per block
func ReadPerBlock(v int64) Option {
	return func(c *Options) { c.ReadPerBlock = v }
}

// HandleTime determines how long each message handling may take, including block time
func HandleTime(v time.Duration) Option {
	return func(c *Options) { c.HandleTime = v }
}

// FromOptions if options are parsed from the environment manually
func FromOptions(v *Options) Option {
	return func(c *Options) { *c = *v }
}
