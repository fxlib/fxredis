package stream

import (
	"time"

	"github.com/caarlos0/env/v6"
)

// ConsumerConf configures a Redis consumer
type ConsumerConf struct {
	GroupName    string        `env:"GROUP_NAME,required"`
	StreamNames  []string      `env:"STREAM_NAMES"`
	StreamStart  int           `env:"STREAM_START"`
	BlockTime    time.Duration `env:"BLOCK_TIME" envDefault:"1s"`
	ReadPerBlock int64         `env:"READ_PER_BLOCK" envDefault:"2"`
	HandleTime   time.Duration `env:"HANDLE_TIME" envDefault:"10s"`
}

// ParseConsumerEnv parses env for consumer config
func ParseConsumerEnv(prefix ...string) (cfg ConsumerConf, err error) {
	var eopts env.Options
	if len(prefix) > 0 {
		eopts.Prefix = string(prefix[0])
	}
	return cfg, env.Parse(&cfg, eopts)
}
