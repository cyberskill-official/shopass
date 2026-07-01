package config

type ScrapeConfig struct {
	MaxConcurrency map[int16]int `yaml:"max_concurrency"`
	MaxAttempts    int           `yaml:"max_attempts"`
	BackoffBaseMs  int           `yaml:"backoff_base_ms"`
	Proxy          ProxyConfig   `yaml:"proxy"`
}

type ProxyConfig struct {
	DailyBudgetMicro int64                  `yaml:"daily_budget_micro"`
	Providers        map[string]ProviderCfg `yaml:"providers"`
}

type ProviderCfg struct {
	Tier string `yaml:"tier"`
}
