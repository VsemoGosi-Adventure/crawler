package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type ProducerConfig struct {
	Brokers []string `env:"REDPANDA_BROKER" envSeparator:"," envDefault:"localhost:9092"`
	Topic   string   `env:"CRAWLER_TOPIC" envDefault:"events.raw"`
}

type Config struct {
	ProducerConfig  ProducerConfig
	CrawlerInterval time.Duration `env:"CRAWLER_INTERVAL" envDefault:"30m"`
	TimepadCity     string        `env:"TIMEPAD_CITY"`
	TimepadToken    string        `env:"TIMEPAD_TOKEN"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
