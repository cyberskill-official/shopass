package config

type ScrapeConfig struct {
	MaxConcurrency map[int16]int `yaml:"max_concurrency"`
	MaxAttempts    int           `yaml:"max_attempts"`
	BackoffBaseMs  int           `yaml:"backoff_base_ms"`
}
