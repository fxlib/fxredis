package stream

import "github.com/caarlos0/env/v6"

// ConsumerConf configures a Redis consumer
type ConsumerConf struct {
	GroupName string `env:"GROUP_NAME"`
}

// ParseConsumerEnv parses env for consumer config
func ParseConsumerEnv(prefix ...string) (cfg ConsumerConf, err error) {
	var eopts env.Options
	if len(prefix) > 0 {
		eopts.Prefix = string(prefix[0])
	}
	return cfg, env.Parse(&cfg, eopts)
}
