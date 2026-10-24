package config

import "github.com/caarlos0/env/v11"

type ProducerConfig struct {
}

type Config struct {
	ProducerConfig ProducerConfig
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
