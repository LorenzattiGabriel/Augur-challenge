package config

import (
	"github.com/caarlos0/env/v10"
)

type Config struct {
	ServerPort     string `env:"SERVER_PORT" envDefault:"8080"`
	ServerHost     string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	DBHost         string `env:"DB_HOST" envDefault:"localhost"`
	DBPort         string `env:"DB_PORT" envDefault:"5432"`
	DBUser         string `env:"DB_USER" envDefault:"postgres"`
	DBPassword     string `env:"DB_PASSWORD" envDefault:"postgres"`
	DBName         string `env:"DB_NAME" envDefault:"threat_intel"`
	DBSSLMode      string `env:"DB_SSLMODE" envDefault:"disable"`
	CacheMaxSizeMB int64  `env:"CACHE_MAX_SIZE_MB" envDefault:"100"`
	RateLimitRPM   int    `env:"RATE_LIMIT_RPM" envDefault:"100"`
	Environment    string `env:"APP_ENV" envDefault:"development"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) DatabaseURL() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}
