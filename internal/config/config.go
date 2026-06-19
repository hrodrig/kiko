package config

import (
	"fmt"

	"github.com/hrodrig/kiko/internal/log"
	"github.com/spf13/viper"
)

type Config struct {
	Listen       string       `mapstructure:"listen"`
	PublicURL    string       `mapstructure:"public_url"`
	LogLevel     string       `mapstructure:"log_level"`
	Database     DatabaseCfg  `mapstructure:"database"`
	Buffer       BufferCfg    `mapstructure:"buffer"`
	RateLimit    RateLimitCfg `mapstructure:"rate_limit"`
	AllowedHosts []string     `mapstructure:"allowed_hosts"`
	Log          *log.Logger
}

type DatabaseCfg struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`

	// DSN overrides all above if set
	DSN string `mapstructure:"dsn"`
}

func (d DatabaseCfg) DSNString() string {
	if d.DSN != "" {
		return d.DSN
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

type BufferCfg struct {
	FlushInterval int `mapstructure:"flush_interval"` // seconds
	Capacity      int `mapstructure:"capacity"`       // max hits in channel
}

type RateLimitCfg struct {
	RequestsPerSec int `mapstructure:"requests_per_sec"`
	Burst          int `mapstructure:"burst"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("kiko")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/kiko/")
	v.SetEnvPrefix("KIKO")
	v.AutomaticEnv()

	// bind env vars for database
	v.BindEnv("database.host", "KIKO_DATABASE_HOST")
	v.BindEnv("database.port", "KIKO_DATABASE_PORT")
	v.BindEnv("database.user", "KIKO_DATABASE_USER")
	v.BindEnv("database.password", "KIKO_DATABASE_PASSWORD")
	v.BindEnv("database.dbname", "KIKO_DATABASE_DBNAME")
	v.BindEnv("database.sslmode", "KIKO_DATABASE_SSLMODE")
	v.BindEnv("database.dsn", "KIKO_DATABASE_DSN")

	v.SetDefault("listen", ":8080")
	v.SetDefault("public_url", "http://localhost:8080")
	v.SetDefault("log_level", "info")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "kiko")
	v.SetDefault("database.password", "")
	v.SetDefault("database.dbname", "kiko")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("buffer.flush_interval", 10)
	v.SetDefault("buffer.capacity", 4096)
	v.SetDefault("rate_limit.requests_per_sec", 100)
	v.SetDefault("rate_limit.burst", 200)

	if path != "" {
		v.SetConfigFile(path)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	cfg.Log = log.New(nil, lvl)

	return &cfg, nil
}
